# nydus

## Initial step

### install dapr

```sh
kubectl config use-context docker-desktop
dapr init -k --runtime-version 1.0.0-rc.3
```

### check work done

```sh
dapr status -k
```

### set redis

redis is for state store and pubsub. 

```sh
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install redis bitnami/redis
k apply -f componants/stateStore.yaml
k apply -f componants/pubsub.yaml
```

# Component Overview

* external: request system
* enter: nydus-enter is http asyncer.
* messagebus: message queue in nydus system.
* callback: return to response to enter when scale out.
* sender: convert from pubsub to request end return to pubsub.
* eventlogger: recode request and response.
* target: target system

```mermaid
sequenceDiagram
  autonumber
  external ->> enter: http request
  activate enter
  enter ->> messagebus: message publish
  Note right of enter: hostIP, targetName required
  messagebus ->> eventlogger: subscribe topic
  messagebus ->> sender: subscribe topic
  activate sender
  sender ->>+ target: message request
  Note right of sender: use targetName
  target ->>- sender: return response
  sender ->> messagebus: response publish
  deactivate sender
  messagebus ->> eventlogger: subscribe response
  messagebus ->> callback: subscribe response
  activate callback
  callback ->> enter: return response
  deactivate callback
  Note left of callback: use hostIP
  enter ->> external: return response
  deactivate enter
```

