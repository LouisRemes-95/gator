Prerequisites:
- Postgres
- Go 1.21+

Once the repo downloaded, use:
go install github.com/LouisRemes-95/gator@latest

Create a ~/.gatorconfig.json (mac/linux) file in home directory with:
{
  "db_url": "postgres://<username>:@localhost:5432/gator?sslmode=disable",
  "current_user_name": "Louis"
}

replace <username> with the actual user

Up migrate 5 times with:
goose -dir sql/schema postgres "postgres://<username>:@localhost:5432/gator" up