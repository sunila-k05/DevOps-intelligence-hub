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

    /* ---------------------- CHECKOUT ---------------------- */
    stage('Checkout') {
      steps {
        checkout scm
        script {
          env.GIT_COMMIT = sh(
            returnStdout: true,
            script: 'git rev-parse --short HEAD'
          ).trim()
        }
      }
    }

    /* ---------------------- BUILD IMAGES ---------------------- */
    stage('Build images') {
      steps {
        script {

          /* ---- BACKEND ---- */
          if (fileExists('backend/Dockerfile')) {
            echo "Building backend image..."
            sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} backend"
          } else {
            echo "No backend/Dockerfile found — skipping backend build"
          }

          /* ---- FRONTEND ---- */
          if (fileExists('frontend/Dockerfile') || fileExists('frontend')) {

            if (fileExists('frontend/Dockerfile')) {
              sh "docker build -t ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} frontend"

            } else if (fileExists('frontend/package.json')) {

              echo "Building frontend using node, packaging into nginx..."
              sh '''
                cd frontend
                docker run --rm -v "$PWD":/work -w /work node:18-bullseye sh -c "npm ci --silent && npm run build"
                cd ..

                docker build -t ''' + "${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}" + ''' - <<EOF
                FROM nginx:stable-alpine
                COPY frontend/build/ /usr/share/nginx/html/
                EOF
              '''

            } else {
              echo "Frontend folder exists but no build config — skipping"
            }

          } else {
            echo "No frontend folder — skipping"
          }
        }
      }
    }

    /* ---------------------- DOCKER HUB LOGIN ---------------------- */
    stage('Docker Hub Login') {
      steps {
        withCredentials([
          usernamePassword(
            credentialsId: 'dockerhub',
            usernameVariable: 'DOCKERHUB_USER',
            passwordVariable: 'DOCKERHUB_PASS'
          )
        ]) {
          sh 'echo "$DOCKERHUB_PASS" | docker login -u "$DOCKERHUB_USER" --password-stdin'
        }
      }
    }

    /* ---------------------- PUSH IMAGES ---------------------- */
    stage('Push images') {
      steps {
        script {

          if (sh(returnStatus: true,
            script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-backend:${IMAGE_TAG}"
          } else {
            echo "Backend image not found — skipping push"
          }

          if (sh(returnStatus: true,
            script: "docker image inspect ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG} > /dev/null 2>&1") == 0) {
            sh "docker push ${DOCKERHUB_NAMESPACE}/devops-intel-frontend:${IMAGE_TAG}"
          } else {
            echo "Frontend image not found — skipping push"
          }

        }
      }
    }

    /* ---------------------- OPTIONAL EC2 DEPLOY ---------------------- */
    stage('Optional: Deploy on host (pull & restart)') {
      steps {
        script {
          sh '''
            echo "Deployment stage is optional.  
            Uncomment docker pull & docker compose lines to activate."
          '''
        }
      }
    }
  }

  /* ---------------------- POST ACTIONS (FIXED) ---------------------- */
  post {
    success {
      echo "Build ${env.BUILD_NUMBER}: images pushed to Docker Hub (${DOCKERHUB_NAMESPACE})"
    }
    failure {
      echo "Build failed — check console output"
    }
    always {
      echo "Pipeline finished."   // REQUIRED → fixes the error
    }
  }
}
