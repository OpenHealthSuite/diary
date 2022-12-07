helm upgrade openfooddiary-app ./chart \
  --namespace openfooddiary \
  --create-namespace \
  --install \
  --atomic