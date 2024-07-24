package dashboard

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
)

func GetDashboardCounts(ctx context.Context, db *database.DB) (*model.GetDashboardData, error) {

	var (
		triviaColl    = db.GetCollection("trivia")
		categoryColl  = db.GetCollection("category")
		templatesColl = db.GetCollection("template")
	)

	triviaFilter := bson.M{"isDeleted": false}
	categoryFilter := bson.M{"isDeleted": false}
	templateFilter := bson.M{"isDeleted": false}

	triviaCount, err1 := triviaColl.CountDocuments(ctx, triviaFilter)
	categoryCount, err2 := categoryColl.CountDocuments(ctx, categoryFilter)
	templateCount, err3 := templatesColl.CountDocuments(ctx, templateFilter)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil, gqlerror.Errorf("err while fetching documents")
	}

	return &model.GetDashboardData{
		TemplatesCount: int(templateCount),
		CategoryCount:  int(categoryCount),
		TotalTrivia:    int(triviaCount),
	}, nil
}
