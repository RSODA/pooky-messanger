package chat

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// Conversations
	conversationsTableName       = "conversations"
	conversationsIDColumn        = "id"
	conversationsCreatedAtColumn = "created_at"

	// ConversatioMember
	conversationMemberTableName    = "conversation_members"
	conversationMemberIDColumn     = "conversation_id"
	conversationMemberUserIDColumn = "user_id"

	// Messages
	messagesTableName            = "messages"
	messagesIDColumn             = "id"
	messagesConversationIDColumn = "conversation_id"
	messagesSenderIDColumn       = "sender_id"
	messagesContetColumn         = "content"
	messagesCreatedAtColumn      = "created_at"

	// Users
	usersTableName       = "users"
	usersIDColumn        = "id"
	usersUsernameColumn  = "username"
	usersCreatedAtColumn = "created_at"
)

type ChatRepository interface {
	CreateConversation(ctx context.Context, userID, targetID string) (string, error)
	GetConversationMembers(ctx context.Context, conversationID string) ([]string, error)
	SaveMessage(ctx context.Context, req *SendMessageRequest) error
	GetMessages(ctx context.Context, req *GetMessageRequest) ([]*GetMessageResponse, error)
	GetUserByUsername(ctx context.Context, username string) (string, error)
	GetConversations(ctx context.Context, userID string) ([]string, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) ChatRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateConversation(ctx context.Context, userID, targetID string) (string, error) {
	var id uuid.UUID

	p, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: "read committed",
	})

	if err != nil {
		return "", err
	}
	defer p.Rollback(ctx)

	builder := squirrel.Insert(conversationsTableName).Columns(conversationsIDColumn).Values(uuid.New()).PlaceholderFormat(squirrel.Dollar).Suffix("RETURNING id")
	query, args, err := builder.ToSql()
	if err != nil {
		return "", ErrCreateQuery
	}

	err = p.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", err
	}

	builder = squirrel.Insert(conversationMemberTableName).Columns(conversationMemberIDColumn, conversationMemberUserIDColumn).Values(id, userID).PlaceholderFormat(squirrel.Dollar)
	query, args, err = builder.ToSql()
	if err != nil {
		return "", ErrCreateQuery
	}

	_, err = p.Exec(ctx, query, args...)
	if err != nil {
		return "", err
	}

	builder = squirrel.Insert(conversationMemberTableName).
		Columns(conversationMemberIDColumn, conversationMemberUserIDColumn).
		Values(id, targetID).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err = builder.ToSql()
	if err != nil {
		return "", ErrCreateQuery
	}

	_, err = p.Exec(ctx, query, args...)
	if err != nil {
		return "", err
	}

	err = p.Commit(ctx)
	if err != nil {
		return "", err
	}

	return id.String(), err
}

func (r *repository) GetConversationMembers(ctx context.Context, conversationID string) ([]string, error) {
	builder := squirrel.Select(conversationMemberUserIDColumn).From(conversationMemberTableName).Where(squirrel.Eq{conversationMemberIDColumn: conversationID}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var member uuid.UUID
		if err := rows.Scan(&member); err != nil {
			return nil, err
		}
		members = append(members, member.String())
	}

	return members, nil
}

func (r *repository) SaveMessage(ctx context.Context, req *SendMessageRequest) error {
	builder := squirrel.Insert(messagesTableName).Columns(messagesIDColumn, messagesConversationIDColumn, messagesContetColumn, messagesSenderIDColumn).
		Values(uuid.New(), req.ConversationId, req.Content, req.SenderID).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetMessages(ctx context.Context, req *GetMessageRequest) ([]*GetMessageResponse, error) {
	builder := squirrel.Select(messagesIDColumn, messagesConversationIDColumn, messagesSenderIDColumn, messagesContetColumn).
		From(messagesTableName).
		Where(squirrel.Eq{messagesConversationIDColumn: req.ConversationId}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, ErrCreateQuery
	}

	rows, err := r.db.Query(ctx, query, args...)

	defer rows.Close()

	var messages []*GetMessageResponse
	for rows.Next() {
		var message GetMessageResponse
		if err := rows.Scan(&message.ID, &message.ConversationId, &message.SenderID, &message.Content); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (string, error) {
	builder := squirrel.Select(usersIDColumn).
		From(usersTableName).
		Where(squirrel.Eq{usersUsernameColumn: username}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return "", ErrCreateQuery
	}

	var id string
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *repository) GetConversations(ctx context.Context, userID string) ([]string, error) {
	builder := squirrel.Select(conversationMemberIDColumn).
		From(conversationMemberTableName).
		Where(squirrel.Eq{conversationMemberUserIDColumn: userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, ErrCreateQuery
	}

	var conversations []string
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var conversation string
		if err := rows.Scan(&conversation); err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}
