docker start openfooddiary-cassandra || docker run --name openfooddiary-cassandra \
    -e CASSANDRA_CLUSTER_NAME="OpenFoodDiary Test Cluster" \
    -p 9042:9042 \
    -d \
    docker.io/cassandra