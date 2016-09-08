Prototype repo for mixologist
-----------------------------
-----------------------------

1. Build uses glide for dependency management
2. Dependencies are vendored
3. Makefile for building

Dev Environment
---------------
1. Install go
2. Install glide (version 0.11.1 only)
3. run: glide install

For Test Deploy
---------------
1. run: sudo apt-get install python-pip
2. run: sudo pip install -r DEMO/requirements.txt
3. Ensure that you have gcloud installed.
4. export PROJECT_ID=mixologist-142215
5. run: DEMO/gcloudinit.py


Update code and redeploy
---------------------------
2. export NAMESPACE=
3. export PROJECT_ID=
4. This step only needs to be done 1 time.
  run: make dev-deploy

5. Make code changes  --- 
6. run: make dev-redeploy
  This assumes you have already run dev-deploy


Notes
-----
1. The demo uses a slightly modified version of the bookstore app
   DEMO/bookstore
2. To build
   docker build -t gcr.io/mjog-1470410002014/bookstore-mixologist -f bookstore.Dockerfile .
