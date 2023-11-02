# good-blast-api
Good Blast REST API

Good Blast is a sample tournament API for mobile games. Every tournament starts on midnight and ends next midnight. Users are grouped into chunks and rewarded according to their ranks. A user can claim their latest rewards in the following day. Tournament API also provides endpoints for local & global leaderboards.

## Installation

Run 
```
docker compose up
```
A local DynamoDB will be instantiated and will be populated with sample data for demonstration purposes.

Then, navigate to the [Postman Desktop Agent](https://www.postman.com/downloads/postman-agent/) and open `docs/Good Blast 3 API.postman_collection.json` file to test the API.

### Postman Collection
You can read descriptions of endpoints on welcome page when you click on **Good Blast 3 API** on the left navigation bar. You can also see **example** responses when you expand each endpoint.


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

