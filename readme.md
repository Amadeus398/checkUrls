# CheckUrls

CheckUrls is a microservice for checking states and a certain number of addresses.
The service accepts links for their subsequent verification at a specified interval 
in seconds (by default - 1 time per day) and saves the status code of the response.

## Start gRPC server

Before you start, configure the environment, where:

- Environment for start CheckUrls:

```
SERVERADDRESS    string // address of server
LOGLEVEL         string // loglevel to display logs
```


- Environment for start PostgreSQL server:

```
HOST       string // host 
PORT       string // port
USER       string // user login
PASSWORD   string // user password
DBNAME     string // DB name
SSLMODE    string // sslmode default value "disabled"
```

To get started gRPC server, run
```
$ go run main.go server
```


## Description of the gRPC client operation

In order for the CheckUrls to function correctly, you need to enter
the data in the database tables.
```
Please note that the CheckUrls can only work with PostgreSQL, because 
the pgx driver is installed.
```

Table *Sites* stores url, frequency and deleted of site, for example:

|  | id | url | frequency | deleted |
---|---:|:---|:---|:---|
1| 1 | http://example.com | 20 | false |

Table *Statuses* stores a date, status code and site_id of site, for example:

|  | id | date | status_code | site_id |
---|---:|:---|:---|:---|
1| 1 | 0001-01-01 00:00:00.000000 | 200 | 1 |

*Note that the site_id in the Statuses corresponds to the id in the Sites*

Let's look at the implementation of the gRPC client operations.

To **create** a new site, enter in command line:

```
go run main.go client create <url> <frequency>
```

*Note that the frequency is the specified interval in seconds.
If you don't enter the "frequency", the site will be checked once a day.*

To **read** a specific site, enter in command line:

```
go run main.go client read <site_id>
```

To get **list** of sites, enter in command line:

```
go run main.go client list
```

To **update** a specific site, enter in command line:

```
go run main.go client update update <site_id> <url> <frequency>
```

To **delete** a specific site, enter in command line:

```
go run main.go client delete <site_id>
```

*Note that after the site is deleted, the url availability
will no longer be checked, but the check history will be
saved in database.*

To get **status** of specific site, enter in command line:

```
go run main.go client status <url>
```

*Note that this command returns information about the last 
5 check's of the specified url(the check time and the status 
code of response are displayed)*


## Used libraries

[github.com/jackc/pgx/v4](https://github.com/jackc/pgx)

[github.com/kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)

[github.com/rs/zerolog](https://github.com/rs/zerolog)

[google.golang.org/grpc](https://grpc.io/docs/languages/go/quickstart/)

[google.golang.org/protobuf](https://grpc.io/docs/languages/go/quickstart/)

