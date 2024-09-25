

### Create tinode with mongodb in kubernetes
```shell
# install dependencies
brew install kubectl
brew install minikube

make 
# if something fails run "make clean-pvc" and "make" again

# expose port
minikube service go-app-service

# make curl to address from prev command
curl -X POST -H "Content-Type: application/json" -d '{"userId": "bvsd33dsd3dsd3","email":"user@example.com","username":"user1","password":"password"}' http://YOUR_ADDRESS/signup

# response will be something like:
{"message":"User signed up successfully!","userID":"\"usrd5WdaR6UN3M\""}%  
```