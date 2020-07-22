---
title: Using the API Explorer
weight: 100
---

Harbor integrated swagger UI from 1.8. That means all APIs can be invoked through the Harbor interface. You can navigate to the API Explorer in two ways. 

1. Log in to Harbor and click the "API EXPLORER" button. All APIs will be invoked with the current user's authorization.                         
![navigation bar](../../img/api-explorer-btn.png)

2. Navigate to the Swagger page by using the IP address of your Harbor instance and adding the router "devcenter". For example: https://10.192.111.118/devcenter. Then click the **Authorize** button to give basic authentication to all APIs. All APIs will be invoked with the authorized user's authorization. 
![authentication](../../img/authorize.png)
