This AWS Lambda function that is triggered by [Amazon EC2 Auto Scaling Lifecycle
Hooks](https://docs.aws.amazon.com/autoscaling/ec2/userguide/lifecycle-hooks.html)
and removes Vault nodes from the Raft peer configuration as they are terminated
from the ASG (failure to do so will result in a cluster outage as new nodes are
created and begin to join the cluster).

This Lambda function uses the EC2 instance ID that is received from the
[event](https://docs.aws.amazon.com/autoscaling/ec2/userguide/cloud-watch-events.html#terminate-lifecycle-action).