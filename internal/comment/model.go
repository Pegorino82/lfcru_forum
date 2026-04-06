package comment

import (
	"html/template"
	"time"
)

// Comment — запись комментария в БД.
type Comment struct {
	ID                    int64
	NewsID                int64
	AuthorID              int64
	ParentID              *int64  // nil для корневых комментариев
	ParentAuthorSnapshot  *string // username родителя на момент создания ответа
	ParentContentSnapshot *string // первые 100 рун контента родителя
	Content               string
	CreatedAt             time.Time
}

// CommentView — проекция для отображения (с данными автора и родителя).
type CommentView struct {
	ID             int64
	NewsID         int64
	AuthorID       int64
	AuthorUsername string
	ParentID       *int64  // nil если корневой или родитель удалён (SET NULL)
	ParentAuthor   *string // из snapshot; сохраняется даже при удалении родителя
	ParentSnippet  *string // из snapshot; сохраняется даже при удалении родителя
	Content        string
	ContentHTML    template.HTML // заполняется в сервисе через RenderMentions
	CreatedAt      time.Time
}
