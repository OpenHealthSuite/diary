podman start openfooddiary-neo4j || podman run --name openfooddiary-neo4j \
  -p 7474:7474 \
  -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/s3cr3tly \
  -e NEO4J_apoc_export_file_enabled=true \
  -e NEO4J_apoc_import_file_enabled=true \
  -e NEO4J_apoc_import_file_use__neo4j__config=true \
  -e NEO4JLABS_PLUGINS=\[\"apoc\"\] \
  -d \
  docker.io/neo4j