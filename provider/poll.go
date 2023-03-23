package provider

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/provider/dbms"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (p *provider) GetPoll(ctx context.Context, pollId string) (domain.Poll, error) {
	ctx = logv2.InitFuncContext(ctx)
	dbData, err := p.Dbms.FirestoreDb.GetPoll(ctx, pollId)
	if err != nil || util.IsInterfaceNil(dbData) {
		return domain.Poll{}, err
	}
	var options []domain.Option
	for _, option := range dbData.Options {
		var voters []domain.Voter
		for _, voter := range option.Voters {
			voters = append(voters, domain.Voter{
				UserId: voter.UserId,
			})
		}
		options = append(options, domain.Option{
			Value:  option.Value,
			Weight: option.Weight,
			Voters: voters,
		})
	}
	poll := domain.Poll{
		Id:       dbData.Id,
		Question: dbData.Question,
		Options:  options,
	}
	return poll, nil
}
func (p *provider) UpsertPoll(ctx context.Context, poll domain.Poll) error {
	ctx = logv2.InitFuncContext(ctx)
	err := p.Dbms.FirestoreDb.UpsertPoll(ctx, buildDbmsPoll(poll))
	if err != nil {
		return err
	}
	return nil
}

func buildDbmsPoll(poll domain.Poll) dbms.Poll {
	var options []dbms.Option
	for _, option := range poll.Options {
		var voters []dbms.Voter
		for _, voter := range option.Voters {
			voters = append(voters, dbms.Voter{
				UserId: voter.UserId,
			})
		}
		options = append(options, dbms.Option{
			Value:  option.Value,
			Weight: option.Weight,
			Voters: voters,
		})
	}
	return dbms.Poll{
		Id:       poll.Id,
		Question: poll.Question,
		Options:  options,
	}
}
