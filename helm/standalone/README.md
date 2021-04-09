# helm chart for standalone mode

## example

```
helm repo add nydus https://mrchypark.github.io/nydus/
helm repod update
helm upgrade den nydus/nydus -i --set nydus.pubsub.name="pubsub" --set nydus.pubsub.topic="den" --set nydus.target.root="http://den.qa.sktchatbot.co.kr" 
```