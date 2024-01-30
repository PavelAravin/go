package mongo

import (
	"GoNews/pkg/storage"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db *mongo.Client
}
type Counter struct {
	ID  string
	Seq int
}
type Authors struct {
	ID   int
	Name string
}

func New(constr string) (*Storage, error) {
	mongoOpts := options.Client().ApplyURI(constr)
	client, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		return nil, err
	}
	s := Storage{
		db: client,
	}
	return &s, nil
}

func (s *Storage) Posts() ([]storage.Post, error) {
	var posts []storage.Post
	filter := bson.D{}
	cur, err := s.db.Database("go-news").Collection("posts").Find(context.Background(), filter)
	defer cur.Close(context.Background()) // close the cursor when done

	if err != nil {
		log.Println("Ошибка получения постов:", err.Error()) // handle error
		return nil, err
	}
	for cur.Next(context.Background()) {
		var post storage.Post
		err := cur.Decode(&post)
		if err != nil {
			log.Println("Ошибка декодирования поста:", err.Error()) // handle error
			return nil, err
		}
		//добавим имя автора в post по его post.AuthorID из коллекции authors
		filter := bson.M{"id": post.AuthorID} // получаем автора по ID
		author := s.db.Database("go-news").Collection("authors").FindOne(context.Background(), filter)
		authors := &Authors{}
		author.Decode(authors)
		post.AuthorName = authors.Name
		posts = append(posts, post)

	}
	if cur.Err() != nil {
		log.Println("Ошибка получения постов:", cur.Err().Error()) // handle error
	}
	return posts, cur.Err()

}

func (s *Storage) AddPost(p storage.Post) error {
	p.CreatedAt = time.Now().Unix()
	if !checkAuthorID(p.AuthorID, s.db) {
		log.Println("Ошибка добавления поста, автора с таким id не существует:", p.AuthorID) // handle error
		return nil
	}
	id := getNextSequence(s.db, "posts")
	log.Println(id)
	p.ID = id
	_, err := s.db.Database("go-news").Collection("posts").InsertOne(context.Background(), p)
	if err != nil {
		log.Println("Ошибка добавления поста:", err.Error()) // handle error
		return err
	}
	return nil
}

func (s *Storage) UpdatePost(p storage.Post) error {

	if !checkAuthorID(p.AuthorID, s.db) {
		log.Println("Ошибка обновления поста, автора с таким id не существует:", p.AuthorID) // handle error
		return nil
	}

	filter := bson.M{"id": p.ID} // условие поиска
	_, err := s.db.Database("go-news").Collection("posts").ReplaceOne(context.Background(), filter, p)
	if err != nil {
		log.Println("Ошибка обновления поста:", err.Error()) // handle error
		return err
	}
	return nil

}

func (s *Storage) DeletePost(p storage.Post) error {
	filter := bson.M{"id": p.ID}                                                                   // условие поиска
	_, err := s.db.Database("go-news").Collection("posts").DeleteOne(context.Background(), filter) // удаляем пост
	if err != nil {
		log.Println("Ошибка удаления поста:", err.Error()) // handle error
		return err

	}
	return nil
}

func getNextSequence(client *mongo.Client, name string) int {
	collection := client.Database("go-news").Collection("counters")

	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	updateResult := collection.FindOneAndUpdate(context.TODO(), filter, update, &opt)
	counter := &Counter{}
	updateResult.Decode(counter)

	return counter.Seq
}

func checkAuthorID(id int, client *mongo.Client) bool {
	filterAuthorID := bson.M{"id": id}                                                                    // условие поиска
	cur := client.Database("go-news").Collection("authors").FindOne(context.Background(), filterAuthorID) // получаем коллекцию
	if cur.Err() != nil {
		return false
	}
	authors := &Authors{}
	cur.Decode(authors)
	return true // автор с таким id существует
}
