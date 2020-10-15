package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/rubbenpad/gofood/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

type dgraph struct {
	db *dgo.Dgraph
}

func New() *dgraph {
	host, _ := os.LookupEnv("DGRAPH_HOST")
	dialOptions := append([]grpc.DialOption{},
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)
	connection, err := grpc.Dial(host, dialOptions...)
	if err != nil {
		log.Panic("Couldn't connect to dgraph")
	}

	db := dgo.NewDgraphClient(api.NewDgraphClient(connection))
	return &dgraph{db: db}
}

func (dg *dgraph) Setup() {
	err := dg.db.Alter(context.Background(), &api.Operation{
		Schema: `
			id: string @index(exact) .
			ip: string @index(exact) .
			date: string @index(exact) .
			when: uid @reverse .
			from: uid @reverse .
			owner: uid @reverse .
			products_id: [uid] @reverse .
			name: string .
			age: int .
			price: string .
			device: string .
		`,
	})
	if err != nil {
		log.Panic("Couldn't setup schema to database")
	}
}

func (dg *dgraph) Save(content []byte) error {
	mutation := &api.Mutation{CommitNow: true, SetJson: content}
	_, err := dg.db.NewTxn().Mutate(context.Background(), mutation)
	return err
}

// Make a query to verify if data for a date is already loaded
// and returns a bool accordign the case.
func (dg *dgraph) GetDate(date string) bool {
	variables := map[string]string{"$date": date}
	query := `
		query dateExists($date: string) {
			exists(func: eq(date, $date)) {
				uid date
			}
		}
	`

	res, err := dg.db.NewTxn().QueryWithVars(context.Background(), query, variables)
	if err != nil {
		log.Panic(err)
	}

	type decode struct {
		Exists []domain.Timestamp `json:"exists"`
	}

	exists := decode{}
	jsonErr := json.Unmarshal(res.Json, &exists)
	if jsonErr != nil {
		log.Panic(jsonErr)
	}

	if len(exists.Exists) == 0 {
		return false
	}

	return true
}

func (dg *dgraph) FindTransactions(id string) {
	variables := map[string]string{"$id": id}
	query := `
		query myDemoQuery($id: string) {
			var(func: eq(id, $id)) {
		  		ID as id
		  		~owner {
					products_id { PID as id }
					from { IP as ip }
		  		}
			}
		  
			var(func: uid(PID)) {
		  		~products_id {
					products_id @filter(not uid(PID)) {
			  			SPID as id
					}
		  		}
			}
		
			var(func: uid(ID)) {
		  		transactions: ~owner { TID as id }
			}
		
			history(func: uid(TID)) {
		  		id
		  		device
		  		from { ip }
		  		products_id {
					id
					name
					price
		  		}
			}
		
			IPList(func: uid(IP)) {
		  		uid
		  		ip
		  		buyers: ~from {
					buyer: owner @filter(not uid(ID)) {
			  			id
			  			name
			  			age
					}
		  		}
			}
		  
			suggestions(func: uid(SPID)) {
		  		id
		  		name
		  		price
			}
	  	}
	`

	res, err := dg.db.NewTxn().QueryWithVars(context.Background(), query, variables)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("%s", res)
}