---
title: Developing the Harbor Frontend
---

If you already have a harbor backend environment, you can build a frontend development environment with the following configuration.

1. Create the file proxy.config.json in the directory harbor/src/portal，and config it according to the sample below.

    **NOTE:** You should replace “$IP_ADDRESS” with your own ip address.

    ```json
    {
        "/api/*": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "changeOrigin": true,
            "logLevel": "debug"
        },
        "/service/*": {
            "target": "$IP_ADDRESS",
            "secure": false, 
            "logLevel": "debug"
        },
        "/c/login": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/sign_in": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/c/log_out": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/sendEmail": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/language": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/reset": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/userExists": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/reset_password": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/i18n/lang/*.json": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug",
            "pathRewrite": { "^/src$": "" }
        },
        "/chartrepo": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        },
        "/*.json": {
            "target": "$IP_ADDRESS",
            "secure": false,
            "logLevel": "debug"
        }
    }
    ```

2. Open the terminal and run the following command，install npm packages as 3rd-party dependencies.

    ```sh
    cd harbor/src/portal
    npm install
    ```

3. Execute the following command，serve Harbor locally.

    ```sh
    npm run start
    ```

4. Then you can visit the Harbor by address: https://localhost:4200.
