# helm chart for standalone mode

## example

```
helm repo add nydus https://mrchypark.github.io/nydus/
helm repod update
helm upgrade den nydus/nydus -i \
    --set nydus.pubsub.name="pubsub" \
    --set nydus.pubsub.topic="httpbin" \
    --set nydus.target.root="https://httpbin.org" 
```