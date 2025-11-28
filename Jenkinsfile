pipeline {
    agent any

    environment {
        AWS_REGION = "us-east-1"
        ACCOUNT_ID = "<AWS-ACCOUNT-ID>"
        REPO_NAME = "<ECR-REPO-NAME>"
        IMAGE_TAG = "${BUILD_NUMBER}"
        ECR_URL = "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    }

    stages {

        stage('Checkout Code') {
            steps {
                git branch: 'main', url: 'https://github.com/<your-username>/<your-repo>.git'
            }
        }

        stage('Login to AWS ECR') {
            steps {
                sh """
                aws --version
                aws ecr get-login-password --region ${AWS_REGION} | \
                docker login --username AWS --password-stdin ${ECR_URL}
                """
            }
        }

        stage('Build Docker Image') {
            steps {
                sh "docker build -t ${REPO_NAME}:${IMAGE_TAG} ."
            }
        }

        stage('Tag & Push to ECR') {
            steps {
                sh """
                docker tag ${REPO_NAME}:${IMAGE_TAG} ${ECR_URL}/${REPO_NAME}:${IMAGE_TAG}
                docker push ${ECR_URL}/${REPO_NAME}:${IMAGE_TAG}
                """
            }
        }

        stage('Deploy to ECS') {
            steps {
                sh """
                aws ecs update-service \
                --cluster <ECS-CLUSTER> \
                --service <ECS-SERVICE> \
                --force-new-deployment \
                --region ${AWS_REGION}
                """
            }
        }
    }
}
