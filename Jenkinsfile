pipeline {
  agent any

  environment {
    DOCKERHUB_NAMESPACE = "sunilak05"
    // safer IMAGE_TAG using build number + short commit
    GIT_COMMIT = ""
    IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT}"
    COMPOSE_FILE = "docker-compose.yml"
    DEPLOY_HOST = "3.7.45.192"          // update when IP changes or use DNS
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
          // set env.GIT_COMMIT to short sha
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
            echo "Building backend image..."
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} backend"
          } else {
            echo "No backend/Dockerfile found, skipping backend build"
          }
        }
      }
    }

    stage('Build Frontend') {
      steps {
        script {
          if (fileExists('frontend/Dockerfile')) {
            echo "Building frontend (Dockerfile)..."
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} frontend"
          } else if (fileExists('frontend/package.json')) {
            echo "Building frontend (npm build -> nginx image)..."
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
            echo "No frontend source detected; skipping frontend build"
          }
        }
      }
    }

    stage('DockerHub Login') {
      steps {
        // replace 'dockerhub' with your Jenkins credential id for Docker Hub
        withCredentials([usernamePassword(credentialsId: 'dockerhub', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
          sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'
        }
      }
    }

    stage('Push Images') {
      steps {
        script {
          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}"
          } else {
            echo "Backend image not found locally, skipping push"
          }

          if (sh(returnStatus: true, script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}"
          } else {
            echo "Frontend image not found locally, skipping push"
          }
        }
      }
    }

    stage('Deploy to EC2') {
      steps {
        // replace 'ec2-ssh' with your Jenkins SSH credentials id (private key)
        sshagent(['ec2-ssh']) {
          // use heredoc (EOF) with single-quoted delimiter to avoid interpolation issues
          sh """
            ssh -o StrictHostKeyChecking=no ubuntu@${DEPLOY_HOST} << 'EOF'
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
            EOF
          """
        }
      }
    }
  } // end stages

  post {
    success {
      echo "Build ${env.BUILD_NUMBER}: images pushed and deployment attempted to ${DEPLOY_HOST}"
    }
    failure {
      echo "Build failed — check console output"
    }
    always {
      // optional cleanup on controller (be careful if controller also runs other jobs)
      sh 'docker image prune -af || true'
    }
  }
}
