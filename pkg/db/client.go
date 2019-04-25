package db

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/dgraph-io/badger"
)

// Client wraps operations for badger.
type Client struct {
	db *badger.DB
}

// Item represents kv
type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NewClient create a new client with db path.
func NewClient(path string) (*Client, error) {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "create db client error")
	}
	return &Client{db}, nil
}

// Close badger.
func (c *Client) Close() error {
	return c.db.Close()
}

// Get by key
func (c *Client) Get(key string) (string, error) {
	var b []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		b, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return "", errors.Wrap(err, "get error")
	}
	return string(b), nil
}

// Set key value
func (c *Client) Set(key, value string) error {
	if err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	}); err != nil {
		return errors.Wrap(err, "set error")
	}
	return nil
}

// Delete by key
func (c *Client) Delete(key string) error {
	if err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	}); err != nil {
		return errors.Wrap(err, "delete error")
	}
	return nil
}

// List by prefix
func (c *Client) List(expr string, skip, limit uint) (vals []*Item, total uint, err error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, 0, errors.Wrap(err, "list error")
	}

	skipped := uint(0)
	added := uint(0)
	c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 20
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			if !re.Match(key) {
				// fmt.Printf("%v not match %v\n", string(item.Key()), expr)
				continue
			}
			total++
			if skipped < skip {
				skipped++
				continue
			}
			if added >= limit {
				continue
			}
			val, _ := item.ValueCopy(nil)
			vals = append(vals, &Item{string(key), string(val)})
			added++
		}
		return nil
	})
	return
}
