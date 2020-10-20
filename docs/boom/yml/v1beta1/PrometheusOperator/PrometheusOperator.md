# PrometheusOperator 
 

## Structure 
 

| Attribute    | Description                                                                                   | Default | Collection | Map  |
| ------------ | --------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy       | Flag if tool should be deployed                                                               |  false  |            |      |
| nodeSelector | NodeSelector for deployment                                                                   |         |            | X    |
| tolerations  | Tolerations to run prometheus-operator on nodes , [here](toleration/Toleration/Toleration.md) |         | X          |      |