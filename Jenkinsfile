pipeline {
    agent any

    environment {
        GITHUB = credentials('github')
        DOCKERHUB = credentials('dockerhub-cred')

        DOCKER_BUILDKIT = "1"
        BUILDKIT_STEP_LOG_MAX_SIZE = "104857600"
        BUILDKIT_PROGRESS = "plain"

        DOCKERHUB_USER = "sunilak05"
    }

    stages {

        /* ---------------------- CHECKOUT ---------------------- */
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
                        sh """
                            docker build -t ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG} backend
                        """
                    } else {
                        echo "Backend Dockerfile missing — skipping."
                    }
                }
            }
        }

        /* ---------------------- BUILD FRONTEND ---------------------- */
        stage('Build Frontend') {
            steps {
                script {
                    if (fileExists('frontend/Dockerfile')) {
                        sh """
                            docker build -t ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG} frontend
                        """
                    } else {
                        echo "Frontend Dockerfile missing — skipping."
                    }
                }
            }
        }

        /* ---------------------- DOCKER LOGIN ---------------------- */
        stage('Docker Login') {
            steps {
                sh """
                    echo '${DOCKERHUB_PSW}' | docker login -u ${DOCKERHUB_USR} --password-stdin
                """
            }
        }

        /* ---------------------- PUSH IMAGES ---------------------- */
        stage('Push Images') {
            steps {
                script {
                    sh "docker push ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG}"
                    sh "docker push ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}"
                }
            }
        }

        /* ---------------------- DEPLOY TO EC2 ---------------------- */
        stage('Deploy to EC2') {
            steps {
                sshagent (credentials: ['ec2-ssh']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ubuntu@13.233.102.98 '
                            docker pull ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG}
                            docker pull ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}

                            docker stop devops-backend || true
                            docker rm devops-backend || true

                            docker stop devops-frontend || true
                            docker rm devops-frontend || true

                            docker run -d --name devops-backend -p 8081:8081 ${DOCKERHUB_USER}/devops-intel-backend:${env.TAG}
                            docker run -d --name devops-frontend -p 80:80 ${DOCKERHUB_USER}/devops-intel-frontend:${env.TAG}
                        '
                    """
                }
            }
        }
    }

    post {
        success {
            echo "SUCCESS — deployed version ${env.TAG}"
        }
        failure {
            echo "PIPELINE FAILED — check logs!"
        }
    }
}
