package provider

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/provider/dbms"
)

func (p *provider) HelloWorld(ctx context.Context) error {
	//err := p.Dbms.FirestoreDb.HelloWorld(ctx)
	//if err != nil {
	//	return err
	//}
	ctx = logv2.InitFuncContext(ctx)
	err2 := p.Dbms.FirestoreDb.InsertLogActivity(ctx, dbms.LogActivity{
		UserId:   "id",
		Username: "username",
		Activity: "activity",
	})
	if err2 != nil {
		return err2
	}
	return nil
}
