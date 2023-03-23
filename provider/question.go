package provider

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (p *provider) UpsertLatestQuestion(ctx context.Context, q domain.Question) error {
	ctx = logv2.InitFuncContext(ctx)
	var err error
	if q.DocId != "" {
		err = p.Dbms.FirestoreDb.UpdateQuestionTime(ctx, q.DocId)
	} else {
		err = p.Dbms.FirestoreDb.InsertLatestQuestion(ctx, q.UserId, q.GuildId)
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *provider) GetLatestQuestion(ctx context.Context, userId string, guildId string) (domain.Question, error) {
	ctx = logv2.InitFuncContext(ctx)

	dbData, err := p.Dbms.FirestoreDb.GetLatestQuestion(ctx, userId, guildId)
	if err != nil {
		return domain.Question{}, err
	}

	if util.IsInterfaceNil(dbData) {
		q := domain.Question{
			UserId:  userId,
			GuildId: guildId,
		}
		logv2.Debug(ctx, logv2.Info, q, "No question found, returning new one")
		return q, nil
	}

	q := domain.Question{
		UserId:  dbData.UserId,
		GuildId: dbData.GuildId,
		Time:    dbData.Time,
		DocId:   dbData.DocId,
	}
	return q, nil
}
