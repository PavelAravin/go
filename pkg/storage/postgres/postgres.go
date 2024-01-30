package postgres

import (
	"GoNews/pkg/storage"
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Хранилище данных.
type Storage struct {
	db *pgxpool.Pool
}

// Конструктор, принимает строку подключения к БД.
func New(constr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Storage{
		db: db,
	}
	return &s, nil
}

// Posts возвращает список задач из БД.
func (s *Storage) Posts() ([]storage.Post, error) {
	rows, err := s.db.Query(context.Background(), `
	SELECT 
		p.id AS post_id, 
    	a.name AS author_name,
		a.id as author_id, 
    	p.title, 
    	p.content, 
    	p.created_at 
	FROM 
    	posts p 
	JOIN 
    	authors a ON p.author_id = a.id
	ORDER BY 
		p.id	
	`,
	)
	if err != nil {
		log.Println("Ошибка выборки всех постов")
		log.Println(err)
		return nil, err
	}
	var posts []storage.Post

	for rows.Next() {
		var p storage.Post
		err = rows.Scan(
			&p.ID,
			&p.AuthorName,
			&p.AuthorID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
		)
		if err != nil {
			log.Println("Ошибка выборки всех постов:")
			log.Println(err)
			return nil, err
		}
		// добавление переменной в массив результатов
		posts = append(posts, p)

	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	// ВАЖНО не забыть проверить rows.Err()
	return posts, nil
}

func (s *Storage) AddPost(p storage.Post) error {
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO 
		 	posts (author_id,title,content)
		VALUES 
			($1, $2,$3) 
		RETURNING
		 	id;
		`,
		p.AuthorID,
		p.Title,
		p.Content,
	)
	if err != nil {
		log.Println("Ошибка добавления поста:")
		log.Println(err)
	}
	return err
}
func (s *Storage) UpdatePost(p storage.Post) error {

	_, err := s.db.Exec(context.Background(), `
	UPDATE 
		posts 
	SET 
		title = $1, 
		content = $2, 
		author_id = $3  
	WHERE 
		id = $4;`,
		p.Title,
		p.Content,
		p.AuthorID,
		p.ID)
	if err != nil {
		log.Println("Ошибка обновления поста:")
		log.Println(err)
		return err
	}
	return nil
}
func (s *Storage) DeletePost(p storage.Post) error {
	_, err := s.db.Exec(context.Background(), ` 
	DELETE FROM 
		posts 
	WHERE 
		id = $1`,
		p.ID)
	if err != nil {
		log.Println("Ошибка удаленя поста:")
		log.Println(err)
		return err
	}
	return nil
}
