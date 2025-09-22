# book-manage-demo

## How to build

```bash
cd backend
# you need create `conf.json` in `backend` folder.
vim conf.json

cd ..
docker-compose up --build -d

```

## conf.json

```
{
  "debug": false,
  "serverAddr": ":9001",
  "apiKey": "75df",
  "downloadPath": "/loclpath/download",
  "accounts": [{
    "user": "admin",
    "password": "Knpss4NZY-"
  },
  {
    "user": "user1",
    "password": "waf12-"
  }
    ]
  "mysql": "root:waf12@tcp(localhost:3306)/word?charset=utf8&interpolateParams=True&parseTime=true"
}
```

# Clean useless items

## clean deleted item

```jsx
select count(*) from related_book where deleted_at != 0;
delete from related_book where deleted_at != 0;
```

## clean examples and definitions

First, we need delete all deleted items.

examples:

```jsx

select example.id, example.content, related_book.item_id,related_book.item_type from example
LEFT JOIN related_book on example.id = related_book.item_id
AND related_book.item_type="example"
WHERE example.deleted_at = 0 AND related_book.item_id IS NULL;

DELETE FROM example WHERE id IN (
    SELECT * FROM (
				select example.id  from example
				LEFT JOIN related_book on example.id = related_book.item_id
				AND related_book.item_type="example"
				WHERE example.deleted_at = 0 AND related_book.item_id IS NULL
    ) AS p
);

```

definitions:

```jsx
select definition.id, definition.definition, related_book.item_id, related_book.item_type from definition
LEFT JOIN related_book on definition.id = related_book.item_id
AND related_book.item_type="definition"
WHERE definition.deleted_at = 0 AND related_book.item_id IS NULL;

DELETE FROM definition WHERE id IN (
    SELECT * FROM (
				select definition.id  from definition
				LEFT JOIN related_book on definition.id = related_book.item_id
				AND related_book.item_type="definition"
				WHERE definition.deleted_at = 0 AND related_book.item_id IS NULL
    ) AS p
);
```
