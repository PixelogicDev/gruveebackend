# Grüvee Backend FAQ
Here we will list information about some of the inner workings of Grüvee Backend

### 🔥 Adding New Firebase Functions
This project is ordered around Firebase Functions. Each module inside the `cmd` represents it's own Firebase Function. 

#### Updating Build Script
When adding new functions to this project you will need to also update the build script. This is the process to do that:

1. Add a variable for the function replace path. For example if your new function is called `addCoolPerson` create a replace variable that looks like this: `addCoolPerson=github.com/pixelogicdev/gruveebackend/cmd/addcoolperson=../cmd/addcoolperson`
2. Add a new if statement for `addcoolperon/` directory in the if logic
3. Make sure to add the new `go mod edit -replace $addCoolPerson` to all other functions (for now it's easier this way to make sure we aren't missing anything)

We are looking for a much better way to do this, but for now this allows us to use our local changes when developing.

#### Adding Deployment File & Version File
You will notice in every function folder we have a `.deployment` and `.version` file:

`.deployment` - This file includes the script we need to run in order to deploy this function to the cloud and actually utilize is

`.version` - This file allows us to keep track of the changes in each function and keeps our tags in sync with what is currentyl in master and what is being developed.


### 🔀 Merge Process
Golang goes off of version numbers in Github. In order for our Firebase functions in the cloud to work properly we need to make sure they download the latest version of each of these functions from this repo. When ready to merge new changes into master we need to do the following:

1. Before merging into `master` we need to make sure to go into every function within `/cmd` and remove the replace tags from `go.mod`
2. Then commit the changes and verify things work locally
3. Make sure to update the `.version` file in whatever module the change was made in. Follow the versioning system and add a short description of your changes in that file.
4. Create a Pull Request to merge into `master`
5. Once merged into `master` we need to tag the module that was changed (We are using this format: `v1.0.0-beta.{WhateverNumberComesNext}`.)
6. This should happen on the `master` branch so make sure to pull the latest and start the tagging process
7. The tag needs to happen on the module like so: `cmd/{ModuleName}/{NewVersionNumber} ([Please use this README for reference](https://github.com/go-modules-by-example/index/blob/master/009_submodules/README.md))
8. Once all the tags have been added, use `git push origin --tags` to push all the tags to master
9. Verify that the module is not being used by any other modules. If it is, make sure to make another Pull Request to `master` with this change.

### 🛫 Deploying Firebase Functions
After all the tagging and merging into master is good to go, we are ready to deploy to Firebase! The process for this is as follows:
   1. Change `config.yaml` file `ENVIORNMENT: PROD` 
   2. cd into cmd/function folder
   3. In all the functions that need to be redeployed, head over to the `.deployment` and copy the script. It should look something like this
```
gcloud functions deploy {whatYourHTTPFunctionWillBeCalled} \ 
--entry-point {ActualClassEntryPoint} \ 
--runtime go113 \ 
--trigger-http
--allow-unauthenticated
--env-vars-file ../..internal/config.yaml
```
OR
```
gcloud functions deploy whatYourTriggerFunctionWillBeCalled \
    --runtime go113 \
    --trigger-event providers/cloud.firestore/eventTypes/document.create \
    --trigger-resource "projects/{FirebaseId}/databases/(default)/{PathToCollection}" \
    --env-vars-file ../../internal/config.yaml
```

### 🖌 WIPS

### Test Firebase Function Trigger Events Locally
1. In `internal/helpers/localCloudTrigger` we had an endpoint that creates a new cloud event that points to a specific cloud trigger.
2. Using something like Insomnia, trigger the localCloudTrigger endpoint with a `FirestoreEvent` payload.
3. That gets fired off and will "create" a trigger event for the endpoint you pass to it