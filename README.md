## Handin-3 Fadbamserne
Created by:
Albert Ross Johannessen,
Daniel Fich,
Felix Anton Andersson

### How to run this program
You need docker to be able to run this program, which you can download [here](https://www.docker.com).

When docker has been downloaded you can just run the following command withing the `./Fadbamsernes-Chatservice` folder.
```shell
docker compose up --build
```

If you do not want to use docker you can also just use the command specified in the dockerfiles in `./client/Dockerfile` and `./server/Dockerfile`. Though this is not recommended because communication happends through the docker network, where stuff is extracted through environment variables.

If you want to try scaling the application up with more client you can try changing the `replicas` withing the compose.yaml file, which is currently set to 3.
