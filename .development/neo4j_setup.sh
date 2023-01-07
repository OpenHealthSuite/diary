podman start openfooddiary-neo4j || podman run --name openfooddiary-neo4j \
  -p 7474:7474 \
  -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/s3cr3tly \
  -d \
  docker.io/neo4j