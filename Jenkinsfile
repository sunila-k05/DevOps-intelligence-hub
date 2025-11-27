pipeline {
    agent any

    environment {
        // GitHub + DockerHub credentials
        GITHUB = credentials('github')
        DOCKERHUB = credentials('dockerhub-cred')

        // BuildKit for fast builds
        DOCKER_BUILDKIT = "1"
        BUILDKIT_STEP_LOG_MAX_SIZE = "104857600"
        BUILDKIT_PROGRESS = "plain"

        // DockerHub username
        DOCKERHUB_USER = "sunilak05"

        // Tag placeholder
        TAG = ""
    }

    stages {

        /* ---------------------- CHECKOUT CODE ---------------------- */
        stage('Checkout') {
            steps {
                checkout scm

                script {
                    env.TAG = sh(
                        script: "git rev-parse --short HEAD",
                        returnStdout: true
                    ).trim()

                    echo "Using TAG: ${env.TAG}"
                }
            }
        }

        /* ---------------------- PRELOAD BASE IMAGES ---------------------- */
        stage('Preload Base Images') {
            steps {
                sh """
                    echo "Pre-pulling base images..."
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
                        echo "Building backend..."
                        sh """
                            docker build \
                                -t ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG} \
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
                        echo "Building frontend..."
                        sh """
                            docker build \
                                -t ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG} \
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
                        -u ${DOCKERHUB_USR} --password-stdin
                """
            }
        }

        /* ---------------------- PUSH IMAGES TO DOCKER HUB ---------------------- */
        stage('Push Images') {
            steps {
                script {

                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG}"
                    }

                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}"
                    }
                }
            }
        }

        /* ---------------------- DEPLOY TO EC2 ---------------------- */
        stage('Deploy to EC2') {
            when { expression { true } }  // always deploy
            steps {
                sshagent(credentials: ['ec2-ssh']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ubuntu@13.233.102.98 "
                            echo 'Pulling latest images...' &&

                            docker pull ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG} || true &&
                            docker pull ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG} || true &&

                            echo 'Stopping old containers...' &&
                            docker stop devops-backend || true &&
                            docker stop devops-frontend || true &&

                            echo 'Removing old containers...' &&
                            docker rm devops-backend || true &&
                            docker rm devops-frontend || true &&

                            echo 'Starting backend...' &&
                            docker run -d --name devops-backend -p 8081:8081 \\
                                ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG} &&

                            echo 'Starting frontend...' &&
                            docker run -d --name devops-frontend -p 80:80 \\
                                ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}
                        "
                    """
                }
            }
        }
    }

    post {
        success {
            echo "SUCCESS — images pushed & deployed. Version: ${env.TAG}"
        }
        failure {
            echo "BUILD FAILED — check errors."
        }
    }
}
