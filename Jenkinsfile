library identifier: "pipeline-library@master",
retriever: modernSCM(
  [
    $class: "GitSCMSource",
    remote: "https://github.com/redhat-cop/pipeline-library.git"
  ]
)

openshift.withCluster() {
  env.NAMESPACE = openshift.project()
  env.APP_NAME = "${JOB_NAME}".replaceAll(/-build.*/, '')
  echo "Starting Pipeline for ${APP_NAME}..."
}

pipeline {

  agent {
    label 'maven'
  }

  stages {
  	stage('Git Checkout') {
      steps {
        git url: "${SOURCE_REPOSITORY_URL}"
      }
    }
  }

}