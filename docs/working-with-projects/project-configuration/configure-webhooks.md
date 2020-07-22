---
title: Configure Webhook Notifications
weight: 45
---

If you are a project administrator, you can configure a connection from a project in Harbor to a webhook endpoint. If you configure webhooks, Harbor notifies the webhook endpoint of certain events that occur in the project. Webhooks allow you to integrate Harbor with other tools to streamline continuous integration and development processes. 

The action that is taken upon receiving a notification from a Harbor project depends on your continuous integration and development processes. For example, by configuring Harbor to send a `POST` request to a webhook listener at an endpoint of your choice, you can trigger a build and deployment of an application whenever there is a change to an image in the repository.

### Supported Events

You can define multiple webhook endpoints per project. Harbor supports two kinds of endpoints currently,  `HTTP`  and `SLACK`. Webhook notifications provide information about events in JSON format and are delivered by `HTTP` or `HTTPS POST` to an existing webhhook endpoint URL or Slack address that you provide. The following table describes the events that trigger notifications and the contents of each notification.

|Event|Webhook Event Type|Contents of Notification|
|---|---|---|
|Push artifact to registry|`PUSH_ARTIFACT`|Repository namespace name, repository name, resource URL, tags, manifest digest, artifact name, push time timestamp, username of user who pushed artifact|
|Pull artifact from registry|`PULL_ARTIFACT`|Repository namespace name, repository name, manifest digest, artifact name, pull time timestamp, username of user who pulled artifact|
|Delete artifact from registry|`DELETE_ARTIFACT`|Repository namespace name, repository name, manifest digest, artifact name, artifact size, delete time timestamp, username of user who deleted image|
|Upload Helm chart to chartMuseum|`UPLOAD_CHART`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of push, username of user who uploaded chart|
|Download Helm chart from chartMuseum|`DOWNLOAD_CHART`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of push, username of user who pulled chart|
|Delete Helm chart from chartMuseum|`DELETE_CHART`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of delete, username of user who deleted chart|
|Image scan completed|`SCANNING_COMPLETED`|Repository namespace name, repository name, tag scanned, image name, number of critical issues, number of major issues, number of minor issues, last scan status, scan completion time timestamp, vulnerability information (CVE ID, description, link to CVE, criticality, URL for any fix), username of user who performed scan|
|Image scan failed|`SCANNING_FAILED`|Repository namespace name, repository name, tag scanned, image name, error that occurred, username of user who performed scan|
|Project quota exceeded|`QUOTA_EXCEED`|Repository namespace name, repository name, tags, manifest digest, artifact name, push time timestamp, username of user who pushed artifact|
|Project quota near threshold|`QUOTA_WARNING`|Repository namespace name, repository name, tags, manifest digest, artifact name, push time timestamp, username of user who pushed artifact|
|Artifact replication finished|`REPLICATION`|Repository namespace name, repository name, tags, manifest digest, artifact name, push time timestamp, username of user who trigger the replication|

#### Payload Format

The webhook notification is delivered in JSON format. The following example shows the JSON notification for a push artifact event when using `HTTP` kind endpoint:

```json
{
	"type": "PUSH_ARTIFACT",
	"occur_at": 1586922308,
	"operator": "admin",
	"event_data": {
		"resources": [{
			"digest": "sha256:8a9e9863dbb6e10edb5adfe917c00da84e1700fa76e7ed02476aa6e6fb8ee0d8",
			"tag": "latest",
			"resource_url": "hub.harbor.com/test-webhook/debian:latest"
		}],
		"repository": {
			"date_created": 1586922308,
			"name": "debian",
			"namespace": "test-webhook",
			"repo_full_name": "test-webhook/debian",
			"repo_type": "private"
		}
	}
}
```
when you select the Slack type, and fill a Slack incoming webhook URL as endpoint, the message you received in Slack will be like,
```json
Harbor webhook events
event_type: PUSH_ARTIFACT
occur_at: April 15th at 11:59 AM
operator: admin
event_data:
{
    "resources": [
        {
            "digest": "sha256:8a9e9863dbb6e10edb5adfe917c00da84e1700fa76e7ed02476aa6e6fb8ee0d8",
            "tag": "latest",
            "resource_url": "hub.harbor.com/test-webhook/debian:latest"
        }
    ],
    "repository": {
        "date_created": 1586922308,
        "name": "debian",
        "namespace": "test-webhook",
        "repo_full_name": "test-webhook/debian",
        "repo_type": "private"
    }
}
```

### Webhook Endpoint Recommendations

There are two kinds of endpoints.  For `HTTP` the endpoint that receives the webhook should ideally have a webhook listener that is capable of interpreting the payload and acting upon the information it contains. For example, running a shell script.

And for Slack endpoint, you should follow the [guide of Slack incoming webhook](https://api.slack.com/messaging/webhooks).

### Example Use Cases

You can configure your continuous integration and development infrastructure so that it performs the following types of operations when it receives a webhook notification from Harbor.

- Artifact push: 
  - Trigger a new build immediately following a push on selected repositories or tags.
  - Notify services or applications that use the artifact that a new artifact is available and pull it.
  - Scan the artifact using Clair.
  - Replicate the artifact to remote registries.
- Image scanning:
  - If a vulnerability is found, rescan the image or replicate it to another registry.
  - If the scan passes, deploy the image.

### Configure Webhooks

1. Log in to the Harbor interface with an account that has at least project administrator privileges.

1. Go to **Projects**, select a project, and select **Webhooks**.

    ![Webhooks option](../../../img/webhooks1.png)

1. Select notify type `HTTP`, so the webhook will be send to a HTTP endpoint.

1. Select events that you want to subscribe.

1. Enter the URL for your webhook endpoint listener.

1. If your webhook listener implements authentication, enter the authentication header. 

1. To implement `HTTPS POST` instead of `HTTP POST`, select the **Verifiy Remote Certficate** check box.

    ![Webhook URL](../../../img/webhooks2.png)

1. Click **Test Endpoint** to make sure that Harbor can connect to the listener.

1. Click **Continue** to create the webhook.

When you have created the webhook, you can click on the arrow at the left end to see the status of the different notifications and the  timestamp of the last time each notification was triggered. You can also manage the webhook by clicking the drop list button of `ACTION...` . 

You can modify the webhook, you can also `Enable` or `Disable` the webhook.

![Webhook Status](../../../img/webhooks3.png)

If a webhook notification fails to send, or if it receives an HTTP error response with a code other than `2xx`, the notification is re-sent based on the configuration that you set in `harbor.yml`. 

### Globally Enable and Disable Webhooks

As a Harbor system administrator, you can enable and disable webhook notifications for all projects.

1. Go to **Configuration** > **System Settings**.
1. Scroll down and check or uncheck the **Webhooks enabled** check box.

    ![Enable/disable webhooks](../../../img/webhooks4.png)
