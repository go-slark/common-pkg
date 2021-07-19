package official

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func SetUpdateTime(update bson.M) bson.M {
	set, ok := update["$set"]
	if ok {
		set.(bson.M)["update_time"] = time.Now()
	} else {
		update["$set"] = bson.M{"update_time": time.Now()}
	}

	return update
}

type QueryOptions struct {
	Skip     int64
	Limit    int64
	Sort     bson.M
	//Selector bson.M
}

func ApplyQueryOpts(opts ...QueryOpt) *options.FindOptions {
	query := &options.FindOptions{}
	qo := &QueryOptions{}
	for _, opt := range opts {
		opt(qo)
	}
	if qo.Sort != nil {
		query = query.SetSort(qo.Sort)
	}
	if qo.Skip != 0 {
		query = query.SetSkip(qo.Skip)
	}
	if qo.Limit != 0 {
		query = query.SetLimit(qo.Limit)
	}

	//if qo.Selector != nil {
	//	query = query.SetProjection(qo.Selector)
	//}

	return query
}

type QueryOpt func(*QueryOptions)

func Skip(skip int64) QueryOpt {
	return func(opts *QueryOptions) {
		opts.Skip = skip
	}
}

func Limit(limit int64) QueryOpt {
	return func(opts *QueryOptions) {
		opts.Limit = limit
	}
}

func Sort(fields ...string) QueryOpt {
	return func(opts *QueryOptions) {
		opts.Sort = append(opts.Sort, fields...)
	}
}

func Select(selector bson.M) QueryOpt {
	return func(opts *QueryOptions) {
		opts.Selector = selector
	}
}
