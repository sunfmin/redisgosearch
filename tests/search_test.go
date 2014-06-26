package tests

import (
	"testing"
	"time"

	"github.com/sunfmin/redisgosearch"
	"github.com/sunfmin/mgodb"
	"labix.org/v2/mgo/bson"
)

type Entry struct {
	ID          bson.ObjectId `bson:"_id"`
	GroupID     string
	Title       string
	Content     string
	Attachments []*Attachment
	CreatedAt   time.Time
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

func (attachment *IndexedAttachment) IndexPieces() (r []string, ais []redisgosearch.Indexable) {
	r = append(r, attachment.Attachment.Filename)
	return
}

func (attachment *IndexedAttachment) IndexEntity() (indexType string, key string, entity interface{}, rank int64) {
	key = attachment.Entry.ID.Hex() + attachment.Attachment.Filename
	indexType = "files"
	entity = attachment
	rank = attachment.Entry.CreatedAt.UnixNano()
	return
}

func (attachment *IndexedAttachment) IndexFilters() (r map[string]string) {
	r = make(map[string]string)
	r["group"] = attachment.Entry.GroupID
	return
}

func (entry *Entry) MakeID() interface{} {
	if entry.ID == "" {
		entry.ID = bson.NewObjectId()
	}
	return entry.ID
}

func (entry *Entry) IndexPieces() (r []string, ais []redisgosearch.Indexable) {
	r = append(r, entry.Title)
	r = append(r, entry.Content)

	for _, a := range entry.Attachments {
		r = append(r, a.Filename)
		ais = append(ais, &IndexedAttachment{entry, a})
	}

	return
}

func (entry *Entry) IndexEntity() (indexType string, key string, entity interface{}, rank int64) {
	key = entry.ID.Hex()
	indexType = "entries"
	entity = entry
	rank = entry.CreatedAt.UnixNano()
	return
}

func (entry *Entry) IndexFilters() (r map[string]string) {
	r = make(map[string]string)
	r["group"] = entry.GroupID
	return
}

func TestIndexAndSearch(t *testing.T) {

	mgodb.Setup("localhost", "redisgosearch")

	client := redisgosearch.NewClient("localhost:6379", "theplant")

	e1 := &Entry{
		ID:      bson.ObjectIdHex("50344415ff3a8aa694000001"),
		GroupID: "Qortex",
		Title:   "Thread Safety",
		Content: "The connection http://google.com Send and Flush methods cannot be called concurrently with other calls to these methods. The connection Receive method cannot be called concurrently with other calls to Receive. Because the connection Do method uses Send, Flush and Receive, the Do method cannot be called concurrently with Send, Flush, Receive or Do. Unless stated otherwise, all other concurrent access is allowed.",
		Attachments: []*Attachment{
			{
				Filename:    "QORTEX UI 0.88.pdf",
				ContentType: "application/pdf",
				CreatedAt:   time.Now(),
			},
		},
		CreatedAt: time.Unix(10000, 0),
	}
	e2 := &Entry{
		ID:      bson.ObjectIdHex("50344415ff3a8aa694000002"),
		GroupID: "ASICS",
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
		CreatedAt: time.Unix(20000, 0),
	}

	mgodb.Save("entries", e1)
	client.Index(e1)

	mgodb.Save("entries", e2)
	client.Index(e2)

	var entries []*Entry
	count, err := client.Search("entries", "concurrent access", nil, 0, 10, &entries)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error(entries)
	}
	if entries[0].Title != "Thread Safety" {
		t.Error(entries[0])
	}

	var attachments []*IndexedAttachment
	_, err = client.Search("files", "alternate qortex", map[string]string{"group": "ASICS"}, 0, 20, &attachments)
	if err != nil {
		t.Error(err)
	}

	if attachments[0].Attachment.Filename != "Alternate Qortex Logo.jpg" || len(attachments) != 1 {
		t.Error(attachments[0])
	}

	// sort
	var sorted []*Entry
	client.Search("entries", "other", nil, 0, 10, &sorted)
	if sorted[0].ID.Hex() != "50344415ff3a8aa694000002" {
		t.Error(sorted[0])
	}
}
