package tests

import (
	"github.com/sunfmin/redisgosearch"
	"github.com/sunfmin/mgodb"
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

type Entry struct {
	Id          bson.ObjectId `bson:"_id"`
	Title       string
	Content     string
	Attachments []*Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	CreatedAt   time.Time
}

type IndexedAttachment struct {
	Entry      *Entry
	Attachment *Attachment
}

func (this *IndexedAttachment) IndexPieces() (r []string, ais []redisgosearch.Indexable) {
	r = append(r, this.Attachment.Filename)
	return
}

func (this *IndexedAttachment) IndexEntity() (key string, indexType string, entity interface{}) {
	key = this.Entry.Id.Hex() + this.Attachment.Filename
	indexType = "files"
	entity = this
	return
}

func (this *Entry) MakeId() interface{} {
	if this.Id == "" {
		this.Id = bson.NewObjectId()
	}
	return this.Id
}

func (this *Entry) IndexPieces() (r []string, ais []redisgosearch.Indexable) {
	r = append(r, this.Title)
	r = append(r, this.Content)

	for _, a := range this.Attachments {
		r = append(r, a.Filename)
		ais = append(ais, &IndexedAttachment{this, a})
	}

	return
}

func (this *Entry) IndexEntity() (key string, indexType string, entity interface{}) {
	indexType = "entries"
	key = this.Id.Hex()
	entity = this
	return
}

func TestIndexAndSearch(t *testing.T) {

	mgodb.Setup("localhost", "redisgosearch")

	client := redisgosearch.NewClient("localhost:6379", "theplant")

	e1 := &Entry{
		Id:      bson.ObjectIdHex("50344415ff3a8aa694000001"),
		Title:   "Thread Safety",
		Content: "The connection http://google.com Send and Flush methods cannot be called concurrently with other calls to these methods. The connection Receive method cannot be called concurrently with other calls to Receive. Because the connection Do method uses Send, Flush and Receive, the Do method cannot be called concurrently with Send, Flush, Receive or Do. Unless stated otherwise, all other concurrent access is allowed.",
		Attachments: []*Attachment{
			{
				Filename:    "QORTEX UI 0.88.pdf",
				ContentType: "application/pdf",
				CreatedAt:   time.Now(),
			},
		},
	}
	e2 := &Entry{
		Id:      bson.ObjectIdHex("50344415ff3a8aa694000002"),
		Title:   "redis is a client for the Redis database",
		Content: "The Conn interface is the primary interface for working with Redis. Applications create connections by calling the Dial, DialWithTimeout or NewConn functions. In the future, functions will be added for creating shareded and other types of connections.",
		Attachments: []*Attachment{
			{
				Filename:    "Screen Shot 2012-08-19 at 11.52.51 AM.png",
				ContentType: "image/png",
				CreatedAt:   time.Now(),
			}, {
				Filename:    "Alternate Qortex Logo.jpg",
				ContentType: "image/jpg",
				CreatedAt:   time.Now(),
			},
		},
	}

	mgodb.Save("entries", e1)
	client.Index(e1)

	mgodb.Save("entries", e2)
	client.Index(e2)

	var entries []*Entry
	err := client.Search("entries", "concurrent access", 10, &entries)
	if err != nil {
		t.Error(err)
	}
	if len(entries) != 1 {
		t.Error(entries)
	}
	if entries[0].Title != "Thread Safety" {
		t.Error(entries[0])
	}

	var attachments []*IndexedAttachment
	err = client.Search("files", "alternate qortex", 20, &attachments)
	if err != nil {
		t.Error(err)
	}

	if attachments[0].Attachment.Filename != "Alternate Qortex Logo.jpg" || len(attachments) != 1 {
		t.Error(attachments[0])
	}

}
