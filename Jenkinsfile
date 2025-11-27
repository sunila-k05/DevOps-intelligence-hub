pipeline {
    agent any

    environment {
        // GitHub + DockerHub credentials
        GITHUB = credentials('github')
        DOCKERHUB = credentials('dockerhub-cred')

        // Docker buildkit settings
        DOCKER_BUILDKIT = "1"
        BUILDKIT_STEP_LOG_MAX_SIZE = "104857600"
        BUILDKIT_PROGRESS = "plain"

        // DockerHub username
        DOCKERHUB_USER = "sunilak05"
    }

    stages {

        /* ---------------------- CHECKOUT CODE ---------------------- */
        stage('Checkout') {
            steps {
                checkout scm

                script {
                    TAG = sh(
                        script: "git rev-parse --short HEAD",
                        returnStdout: true
                    ).trim()
                    echo "Using TAG: ${TAG}"
                }
            }
        }

        /* ---------------------- PRELOAD BASE IMAGES ---------------------- */
        stage('Preload Base Images') {
            steps {
                sh """
                    echo "Pre-pulling base images (mirror + cache)..."
                    docker pull registry.hub.docker.com/library/golang:1.22 || true
                    docker pull registry.hub.docker.com/library/alpine:3.20 || true
                    docker pull node:20-alpine || true
                    docker pull nginx:alpine || true
                """
            }
        }

        /* ---------------------- BUILD BACKEND ---------------------- */
        stage('Build Backend') {
            steps {
                script {
                    if (fileExists('backend/Dockerfile')) {
                        echo "Building backend image..."
                        sh """
                            docker build \
                                -t ${DOCKERHUB_USER}/devops-intel-backend:${TAG} \
                                backend
                        """
                    } else {
                        echo "Backend Dockerfile not found — skipping."
                    }
                }
            }
        }

        /* ---------------------- BUILD FRONTEND ---------------------- */
        stage('Build Frontend') {
            steps {
                script {
                    if (fileExists('frontend/Dockerfile')) {
                        echo "Building frontend image..."
                        sh """
                            docker build \
                                -t ${DOCKERHUB_USER}/devops-intel-frontend:${TAG} \
                                frontend
                        """
                    } else {
                        echo "Frontend Dockerfile not found — skipping."
                    }
                }
            }
        }

        /* ---------------------- DOCKER LOGIN ---------------------- */
        stage('Docker Login') {
            steps {
                sh """
                    echo ${DOCKERHUB_PSW} | docker login \
                        -u ${DOCKERHUB_USR} \
                        --password-stdin
                """
            }
        }

        /* ---------------------- PUSH IMAGES ---------------------- */
        stage('Push Images') {
            steps {
                script {

                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-backend:${TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-backend:${TAG}"
                    }

                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}"
                    }
                }
            }
        }

        /* ---------------------- DEPLOY TO EC2 (LATER) ---------------------- */
        stage('Deploy to EC2 (optional)') {
          //when { branch 'main' }
            steps {
                echo "Deployment disabled. Enable later."
            }
        }
    }

    post {
        success {
            echo "SUCCESS — images pushed with tag: ${TAG}"
        }
        failure {
            echo "BUILD FAILED — check errors"
        }
    }
}
