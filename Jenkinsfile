pipeline {
  agent any

  environment {
    DOCKERHUB_NAMESPACE = "sunilak05"
    IMAGE_TAG = "${env.BUILD_NUMBER}-${GIT_COMMIT?.substring(0,7)}"
    COMPOSE_FILE = "docker-compose.yml"
  }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '20'))
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
        script { env.GIT_COMMIT = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim() }
      }
    }

    stage('Build images') {
      steps {
        script {
          // Build backend if present
          if (fileExists('backend/Dockerfile')) {
            echo "Building backend image..."
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} backend"
          } else {
            echo "No backend/Dockerfile found, skipping backend build"
          }

          // Build frontend if present
          if (fileExists('frontend/Dockerfile') || fileExists('frontend')) {
            // If frontend has a Dockerfile, use it; otherwise build static assets and use a simple nginx image
            if (fileExists('frontend/Dockerfile')) {
              sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} frontend"
            } else if (fileExists('frontend/package.json')) {
              echo "Building frontend via node build and packaging into nginx image..."
              sh '''
                cd frontend
                docker run --rm -v "$PWD":/work -w /work node:18-bullseye sh -c "npm ci --silent && npm run build"
                cd ..
                # create a lightweight nginx image with build output
                docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} - <<EOF
                FROM nginx:stable-alpine
                COPY frontend/build/ /usr/share/nginx/html/
                EOF
              '''
            } else {
              echo "No frontend source detected; skipping frontend build"
            }
          } else {
            echo "No frontend folder, skipping frontend build"
          }
        }
      }
    }

    stage('Docker Hub Login') {
      steps {
        withCredentials([usernamePassword(credentialsId: 'dockerhub', usernameVariable: 'DOCKERHUB_USER', passwordVariable: 'DOCKERHUB_PASS')]) {
          sh 'echo "$DOCKERHUB_PASS" | docker login -u "$DOCKERHUB_USER" --password-stdin'
        }
      }
    }

    stage('Push images') {
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

    stage('Optional: Deploy on host (pull & restart)') {
      steps {
        script {
          // If you want to update a docker-compose that references remote images, uncomment the following block.
          // This assumes your docker-compose.yml uses image: sunilak05/devops-intel-frontend:latest (or a template).
          // Here we pull images with the new tag then restart compose.
          sh '''
            # If you want to auto-deploy on the same host after push, uncomment lines below and adjust compose file to use image tags.
            # docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} || true
            # docker pull ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} || true
            # docker compose -f ${COMPOSE_FILE} down || true
            # docker compose -f ${COMPOSE_FILE} up -d
            echo "Deploy step is optional and currently a no-op. Uncomment docker pull/compose lines to activate."
          '''
        }
      }
    }
  }

  post {
    success {
      echo "Build ${env.BUILD_NUMBER}: images pushed to Docker Hub (${DOCKERHUB_NAMESPACE})"
    }
    failure {
      echo "Build failed — check console output"
    }
    always {
      sh 'docker image prune -f || true'
    }
  }
}
