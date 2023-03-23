package provider

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/provider/dbms"
	"time"
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
		RandomMap: map[string]interface{}{
			"keystring":  "valuestring1",
			"keyinteger": 123,
			"keyfloat":   3.14,
			"keyslice":   []string{"valuestring1", "valuestring2"},
			"keymap":     map[string]string{"key": "valuestring2"},
		},
	})
	if err2 != nil {
		return err2
	}
	time.Sleep(1 * time.Second)
	err3 := p.Dbms.FirestoreDb.GetLogActivity(ctx)
	if err3 != nil {
		return err3
	}
	return nil
}
