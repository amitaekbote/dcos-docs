---
post_title: Upgrading
menu_order: 11
---

## Summary

This document provides instructions for upgrading a DC/OS cluster from version 1.7 to 1.8 using AWS cloudformation templates. It is recommended that you familiarize yourself with the [Advanced DC/OS Installation on AWS] before proceeding.

If this upgrade is performed on a supported OS with all prerequisites fulfilled, this upgrade _should_ preserve the state of running tasks on the cluster.

**Important:**

- The [VIP features](/docs/1.8/usage/service-discovery/load-balancing-vips/virtual-ip-addresses/), added in DC/OS 1.8, require that ports 32768 - 65535 are open between all agent and master nodes for both TCP and UDP.
- The DC/OS UI and APIs may be inconsistent or unavailable while masters are being upgraded. Avoid using them until all masters have been upgraded and have rejoined the cluster. You can monitor the health of a master during an upgrade by watching Exhibitor on port 8181.
- Task history in the Mesos UI will not persist through the upgrade.

## Instructions

1. Login to the current leader master of the cluster.
   1. Using DC/OS CLI: 
      ```
      $dcos node ssh --master-proxy --leader
      ```
   1. After you are logged in, run the following command. This command creates a new 1.8 directory /var/lib/dcos/exhibitor/zookeeper as a symlink to the old /var/lib/zookeeper: 
      ```
      $for node in $(dig +short master.mesos); do ssh -o StrictHostKeyChecking=no $node "sudo mkdir -p /var/lib/dcos/exhibitor && sudo ln -s /var/lib/zookeeper /var/lib/dcos/exhibitor/zookeeper"; done
      ```

   1. Go to http://master-node/exhibitor
      * Go to config tab , it should have three fields which have /var/lib/zookeeper 
        ![Exhibitor UI](../img/dcos-exhibitor-fields-before.png)
        ![Exhibitor UI](../img/dcos-exhibitor-fields-before-2.png)
      * Edit the config and change all three fields that contain /var/lib/zookeeper/ to /var/lib/dcos/exhibitor/zookeeper/
        ![Exhibitor UI](../img/dcos-exhibitor-fields-after.png)
        ![Exhibitor UI](../img/dcos-exhibitor-fields-after-2.png)
      * Commit and perform a rolling restart.
      * This will take a couple of minutes and during that time the Exhibitor UI will flash, wait the commit to be performed fully
   1. Make sure the cluster is healthy at this point.
      * Verify you can access the dashboard
      * Verify all components are healthy
   
1. Update Cloudformation stacks.
   1. Generate the new templates following instructions at [Advanced DC/OS Installation Guide][advanced-aws-custom]
   1. See the AWS documentation on updating CloudFormation stacks: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-direct.html 
   1. Updating the zen or infra stack will update the others automatically , if not start with masters then agents
   
1. Deleting instances
   1. Start by deleting older master instances and then delete the older agent instances

## Notes:

[advanced-aws-custom]: 1.8/administration/installing/cloud/aws/advanced/aws-custom/