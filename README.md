# good-blast-api
Good Blast REST API

## Installation

Run `docker compose up` from the root directory. A local DynamoDB will be instantiated and will be populated with sample data for demonstration purposes.

Then, navigate to the [Postman Desktop Agent](https://www.postman.com/downloads/postman-agent/) and open `docs/Good Blast 3 API.postman_collection.json` file to test the API.

> Note: Update `Variables` to see API results for different users/tournaments. 

### Testing

Run tests with
```
docker exec -it good-blast-api-web-1 go test
```

> Note: Docker images must be up and running to run the tests.


## Deployment
The app is deployed in GCP Cloud Run and same endpoints can be accessed by setting `base_url` parameter to [https://good-blast-api-zfbs2ytkgq-lz.a.run.app](https://good-blast-api-zfbs2ytkgq-lz.a.run.app).
There are two jobs on top of the main service:
1. `insert-tournament`: Inserts a record for tomorrow's tournament every day at 6AM.
2. `update-tournament`: Calculates leaderboards for yesterday's tournament every dat at 7AM.
![Deployment](/docs/img/deployment.png)

## Structs
Below is a representation of the structs used in the API. Please refer to legend for better understanding of the structs.

![Structs](/docs/img/structs.png)

