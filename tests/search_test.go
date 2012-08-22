package tests

import (
	"github.com/sunfmin/redisgosearch"
	"github.com/sunfmin/mgodb"
	"labix.org/v2/mgo/bson"
	"testing"
)

type Entry struct {
	Id                  bson.ObjectId `bson:"_id"`
	Title               string
	Content             string
	AttachmentFileNames []string
}

func (this *Entry) MakeId() interface{} {
	if this.Id == "" {
		this.Id = bson.NewObjectId()
	}
	return this.Id
}

func (this *Entry) IndexKey() (r string) {
	r = this.Id.Hex()
	return
}

func (this *Entry) IndexPieces() (r []*redisgosearch.Piece) {
	r = append(r, &redisgosearch.Piece{
		IndexContent: this.Title,
		Type:         "entries",
	})
	r = append(r, &redisgosearch.Piece{
		IndexContent: this.Content,
		Type:         "entries",
	})

	for _, fname := range this.AttachmentFileNames {
		r = append(r, &redisgosearch.Piece{
			IndexContent: fname,
			Type:         "files",
		})
	}

	return
}

func (this *Entry) IndexEntity() (r interface{}) {
	r = this
	return
}

func TestIndexAndSearch(t *testing.T) {

	mgodb.Setup("localhost", "redisgosearch")

	client := redisgosearch.NewClient("localhost:6379", "theplant")

	e1 := &Entry{
		Id:      bson.ObjectIdHex("50344415ff3a8aa694000001"),
		Title:   "Thread Safety",
		Content: "The connection http://google.com Send and Flush methods cannot be called concurrently with other calls to these methods. The connection Receive method cannot be called concurrently with other calls to Receive. Because the connection Do method uses Send, Flush and Receive, the Do method cannot be called concurrently with Send, Flush, Receive or Do. Unless stated otherwise, all other concurrent access is allowed.",
	}
	e2 := &Entry{
		Id:      bson.ObjectIdHex("50344415ff3a8aa694000002"),
		Title:   "redis is a client for the Redis database",
		Content: "The Conn interface is the primary interface for working with Redis. Applications create connections by calling the Dial, DialWithTimeout or NewConn functions. In the future, functions will be added for creating shareded and other types of connections.",
	}

	mgodb.Save("entries", e1)
	client.Index(e1)

	mgodb.Save("entries", e2)
	client.Index(e2)

	r, err := client.SearchInType("concurrent access", "entries")
	if err != nil {
		t.Error(err)
	}
	var entries []*Entry
	r.All(&entries)
	if len(entries) != 1 {
		t.Error(entries)
	}
	if entries[0].Title != "Thread Safety" {
		t.Error(entries[0])
	}

}
