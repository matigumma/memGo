[qdrant-go-client](https://github.com/qdrant/go-client) 

[qdrant](https://qdrant.tech/documentation/quickstart/) service on docker:

```sh
docker run -p 6333:6333 -p 6334:6334 \
    -v $(pwd)/qdrant_storage:/qdrant/storage:z \
    qdrant/qdrant
```