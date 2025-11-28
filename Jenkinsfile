pipeline {
  agent any

  // lightweight global things that are static
  environment {
    DOCKERHUB_NAMESPACE = "sunilak05"
    // keep Docker Compose filename if you later want to use it
    COMPOSE_FILE = "docker-compose.yml"
    // EC2 target (change if needed)
    EC2_USER = "ubuntu"
    EC2_HOST = "13.233.102.98"         // <-- change to your EC2 IP if different
    // credentials IDs used in Jenkins (adjust to the IDs you have)
    DOCKERHUB_CRED_ID = "dockerhub"   // username/password credential in Jenkins
    EC2_SSH_CRED_ID  = "ec2-ssh"      // sshUserPrivateKey credential in Jenkins
  }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '20'))
    // prevent very long-hanging builds
    timeout(time: 60, unit: 'MINUTES')
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
        script {
          // compute a stable image tag: build number + short git sha
          def shortSha = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
          IMAGE_TAG = "${env.BUILD_NUMBER}-${shortSha}"
          echo "IMAGE_TAG -> ${IMAGE_TAG}"
        }
      }
    }

    stage('Pre-pull base images (cache)') {
      steps {
        echo "Pre-pulling common base images to reduce build time..."
        sh '''
          docker pull registry.hub.docker.com/library/golang:1.22 || true
          docker pull registry.hub.docker.com/library/alpine:3.20 || true
          docker pull node:20-alpine || true
          docker pull nginx:alpine || true
        '''
      }
    }

    stage('Build images') {
      steps {
        script {
          // Backend build
          if (fileExists('backend/Dockerfile')) {
            echo "Building backend image..."
            sh """
              docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} backend
            """
          } else {
            echo "No backend/Dockerfile -> skipping backend build"
          }

          // Frontend build
          if (fileExists('frontend/Dockerfile')) {
            echo "Building frontend (Dockerfile) image..."
            sh """
              docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} frontend
            """
          } else if (fileExists('frontend/package.json')) {
            echo "Building frontend (npm build -> packaged into nginx image)..."
            sh '''
              set -o pipefail
              cd frontend
              docker run --rm -v "$PWD":/work -w /work node:18-bullseye sh -c "npm ci --silent && npm run build"
              cd ..
              docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} - <<'EOF'
              FROM nginx:stable-alpine
              COPY frontend/build/ /usr/share/nginx/html/
              EOF
            '''
          } else {
            echo "No frontend sources found -> skipping frontend build"
          }
        }
      }
    }

    stage('DockerHub Login') {
      steps {
        withCredentials([usernamePassword(credentialsId: env.DOCKERHUB_CRED_ID, usernameVariable: 'DOCKERHUB_USR', passwordVariable: 'DOCKERHUB_PSW')]) {
          // use single-quoted sh to avoid Groovy interpolation of secrets (avoids insecure interpolation warning)
          sh('echo "$DOCKERHUB_PSW" | docker login -u "$DOCKERHUB_USR" --password-stdin')
        }
      }
    }

    stage('Push images') {
      steps {
        script {
          // push backend if built
          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            echo "Pushing backend image..."
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}"
          } else {
            echo "Backend image not present locally -> skipping push"
          }

          // push frontend if built
          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            echo "Pushing frontend image..."
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}"
          } else {
            echo "Frontend image not present locally -> skipping push"
          }
        }
      }
    }

    stage('Deploy to EC2 (pull & restart)') {
      steps {
        // use ssh key file injected by withCredentials - no ssh-agent plugin required
        withCredentials([sshUserPrivateKey(credentialsId: env.EC2_SSH_CRED_ID, keyFileVariable: 'EC2_KEYFILE', passphraseVariable: 'EC2_PASSPHRASE', usernameVariable: 'EC2_KEY_USER')]) {
          script {
            // Build the remote commands: pull, stop, remove, run - idempotent
            def remoteCmd = """
              set -e
              echo "Pulling images on remote..."
              docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} || true
              docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} || true

              echo "Stopping any existing containers..."
              docker stop devops-backend || true
              docker stop devops-frontend || true

              echo "Removing old containers..."
              docker rm devops-backend || true
              docker rm devops-frontend || true

              echo "Starting backend..."
              docker run -d --name devops-backend -p 8081:8081 --restart unless-stopped ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} || true

              echo "Starting frontend..."
              docker run -d --name devops-frontend -p 80:80 --restart unless-stopped ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} || true

              echo "Remote deploy complete."
            """

            // If the key has passphrase, ssh will prompt; using -i with the keyfile handles passphrase only if keyfile is unencrypted.
            // If your key has passphrase, either remove passphrase or use ssh-agent (plugin) or provide passphrase handling here.
            sh """
              chmod 600 "${EC2_KEYFILE}"
              ssh -o StrictHostKeyChecking=no -i "${EC2_KEYFILE}" ${EC2_USER}@${EC2_HOST} /bin/bash -l -c '${remoteCmd}'
            """
          }
        }
      }
    }
  }

  post {
    success {
      echo "SUCCESS: images built & (if available) pushed. Deployed tag: ${IMAGE_TAG}"
    }
    failure {
      echo "FAILED — check console output"
    }
    always {
      echo "Pipeline finished."
      // optional cleanup (commented)
      // sh 'docker image prune -f || true'
    }
  }
}
