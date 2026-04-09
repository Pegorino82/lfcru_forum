package forum

import "errors"

var (
	ErrSectionNotFound      = errors.New("раздел не найден")
	ErrTopicNotFound        = errors.New("тема не найдена")
	ErrPostNotFound         = errors.New("сообщение не найдено")
	ErrParentNotFound       = errors.New("сообщение, на которое вы отвечаете, не найдено")
	ErrReplyToReply         = errors.New("нельзя отвечать на ответ")
	ErrEmptyTitle           = errors.New("название не может быть пустым")
	ErrTitleTooLong         = errors.New("название слишком длинное (максимум 255 символов)")
	ErrDescriptionTooLong   = errors.New("описание слишком длинное (максимум 2000 символов)")
	ErrEmptyContent         = errors.New("сообщение не может быть пустым")
	ErrContentTooLong       = errors.New("сообщение слишком длинное (максимум 20000 символов)")
)
