# natshelmresolver
Initializes nats resolver for nats helm chart

# Motivation
When using the NATS resolver approach (https://github.com/nats-io/k8s/tree/main/helm/charts/nats), automating the setup can be challenging. The process requires running nsc, then manually extracting and copying data from various files into the configuration. This approach works for single deployments but complicates automation. While nsc is user-friendly for manual setups, it lacks automation support. However, it does generate the necessary configuration for seamless integration with the NATS Helm chart, enabling automatic merging.
