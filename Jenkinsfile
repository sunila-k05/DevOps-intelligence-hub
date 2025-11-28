pipeline {
  agent any

  environment {
    DOCKERHUB_NAMESPACE = "sunilak05"
    GIT_COMMIT = ""
    IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT}"
    DEPLOY_HOST = "3.7.45.192"      // update when IP changes
  }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '20'))
  }

  stages {

    stage('Checkout') {
      steps {
        checkout scm
        script {
          env.GIT_COMMIT = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
          env.IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT}"
          echo "Using IMAGE_TAG: ${env.IMAGE_TAG}"
        }
      }
    }

    stage('Build Backend') {
      steps {
        script {
          if (fileExists('backend/Dockerfile')) {
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} backend"
          } else {
            echo "No backend Dockerfile"
          }
        }
      }
    }

    stage('Build Frontend') {
      steps {
        script {
          if (fileExists('frontend/Dockerfile')) {
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} frontend"
          } else if (fileExists('frontend/package.json')) {
            sh """
              cd frontend
              docker run --rm -v "\$PWD":/work -w /work node:18-bullseye sh -c "npm ci --silent && npm run build"
              cd ..
              docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} - <<'DOCKERFILE'
FROM nginx:stable-alpine
COPY frontend/dist/ /usr/share/nginx/html/
DOCKERFILE
            """
          } else {
            echo "No frontend detected"
          }
        }
      }
    }

    stage('DockerHub Login') {
      steps {
        withCredentials([usernamePassword(credentialsId: 'dockerhub', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
          sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'
        }
      }
    }

    stage('Push Images') {
      steps {
        script {

          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}"
          }

          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}"
          }
        }
      }
    }

    stage('Deploy to EC2') {
      steps {
        sshagent(['ec2-ssh']) {
          sh """
ssh -o StrictHostKeyChecking=no ubuntu@${DEPLOY_HOST} 'bash -s' <<'DEPLOY_SCRIPT'
set -e

docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} || true
docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} || true

docker stop devops-backend || true
docker rm devops-backend  || true

docker stop devops-frontend || true
docker rm devops-frontend  || true

docker run -d --name devops-backend -p 8081:8081 ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}
docker run -d --name devops-frontend -p 80:80 ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}

echo "Deployment finished on \$(hostname) with tag ${IMAGE_TAG}"
DEPLOY_SCRIPT
          """
        }
      }
    }

  } // <-- closes stages

  post {
    success {
      echo "Build & Deploy SUCCESS"
    }
    failure {
      echo "BUILD FAILED — check logs"
    }
  }

} // <-- closes pipeline
