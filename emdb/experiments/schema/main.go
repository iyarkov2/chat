package main

import (
	"fmt"
	"github.com/iyarkov2/chat/emdb/schema"
)

func main() {
	var memberSeq = schema.Sequence {
		Id  : 3,
		Name: "MemberSeq",
	}

	var users = schema.Collection {
		Name: "User",
		PrimaryKey: schema.PrimaryKey {
			Id  : userPk,
			Name: "UserPK",
			Seq: schema.Sequence {
				Id  : userSeq,
				Name: "UserSeq",
			},
		},
		Indexes: []schema.Index {
			{
				Id: userNameIdx,
				Name: "UserNameIdx",
				Unique: true,
			},
		},

	}
	var groups = schema.Collection {
		PrimaryKey: schema.PrimaryKey {
			Name: "UserPK",
			Seq: schema.Sequence {
				Id  : 1,
				Name: "UserSeq",
			},
		},
	}

	var members = schema.Collection {

	}
	var collections = []schema.Collection {
		users,
		groups,
		members,
	}

	var sequences = []schema.Sequence {
		users.PrimaryKey.Seq,
		groups.PrimaryKey.Seq,
		memberSeq,
	}


	fmt.Printf("Users sequence 1 %s\n", collections[0].PrimaryKey.Seq.Name)
	fmt.Printf("Users seq 1 %s\n", sequences[0].Name)
	sequences[0].Name = "XYZ"
	fmt.Printf("Users sequence 2 %s\n", collections[0].PrimaryKey.Seq.Name)
	fmt.Printf("Users seq 2 %s\n", sequences[0].Name)

	fmt.Println("Schema", schema.NewSchema(collections, sequences))

	fmt.Printf("Users sequence 5 %s\n", collections[0].PrimaryKey.Seq.Name)
	fmt.Printf("Users seq 5 %s\n", sequences[0].Name)
}
