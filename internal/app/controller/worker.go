// Copyright 2020 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/clivern/beetle/internal/app/kubernetes"
	"github.com/clivern/beetle/internal/app/model"
	"github.com/clivern/beetle/internal/app/module"
	"github.com/clivern/beetle/internal/app/util"

	log "github.com/sirupsen/logrus"
)

// Worker controller
func Worker(id int, messages <-chan string) {
	var ok bool
	var err error
	var db module.Database
	var job model.Job
	var cluster *kubernetes.Cluster

	messageObj := model.Message{}
	deploymentRequest := model.DeploymentRequest{}

	log.WithFields(log.Fields{
		"correlation_id": util.GenerateUUID4(),
		"worker_id":      id,
	}).Info(`Worker started`)

	for message := range messages {
		ok, err = messageObj.LoadFromJSON([]byte(message))

		if !ok || err != nil {
			log.WithFields(log.Fields{
				"correlation_id": messageObj.UUID,
				"worker_id":      id,
				"message":        message,
			}).Warn(`Worker received invalid message`)
			continue
		}

		log.WithFields(log.Fields{
			"correlation_id": messageObj.UUID,
			"worker_id":      id,
			"job_id":         messageObj.Job,
		}).Info(`Worker received a new job`)

		db = module.Database{}

		err = db.AutoConnect()

		if err != nil {
			log.WithFields(log.Fields{
				"correlation_id": messageObj.UUID,
				"worker_id":      id,
				"error":          err.Error(),
			}).Error(`Worker unable to connect to database`)
			continue
		}

		defer db.Close()

		job = db.GetJobByID(messageObj.Job)

		ok, err = deploymentRequest.LoadFromJSON([]byte(job.Payload))

		if !ok || err != nil {
			log.WithFields(log.Fields{
				"correlation_id": messageObj.UUID,
				"worker_id":      id,
				"job_id":         messageObj.Job,
				"job_uuid":       job.UUID,
				"error":          err.Error(),
			}).Error(`Invalid job payload`)

			// Job Failed
			now := time.Now()
			job.Status = model.JobFailed
			job.RunAt = &now
			job.Result = fmt.Sprintf("Invalid job payload, UUID %s", messageObj.UUID)
			db.UpdateJobByID(&job)
			continue
		}

		log.WithFields(log.Fields{
			"correlation_id":      messageObj.UUID,
			"worker_id":           id,
			"job_id":              messageObj.Job,
			"job_uuid":            job.UUID,
			"request_cluster":     deploymentRequest.Cluster,
			"request_namespace":   deploymentRequest.Namespace,
			"request_application": deploymentRequest.Application,
			"request_version":     deploymentRequest.Version,
			"request_strategy":    deploymentRequest.Strategy,
		}).Info(`Worker accepted deployment request`)

		cluster, err = kubernetes.GetCluster(deploymentRequest.Cluster)

		if err != nil {
			log.WithFields(log.Fields{
				"correlation_id":      messageObj.UUID,
				"worker_id":           id,
				"error":               err.Error(),
				"request_cluster":     deploymentRequest.Cluster,
				"request_namespace":   deploymentRequest.Namespace,
				"request_application": deploymentRequest.Application,
				"request_version":     deploymentRequest.Version,
				"request_strategy":    deploymentRequest.Strategy,
			}).Error(`Worker can not find the cluster`)

			// Job Failed
			now := time.Now()
			job.Status = model.JobFailed
			job.RunAt = &now
			job.Result = fmt.Sprintf("Worker can not find the cluster, UUID %s", messageObj.UUID)
			db.UpdateJobByID(&job)
			continue
		}

		ok, err = cluster.Ping(context.TODO())

		if !ok || err != nil {
			log.WithFields(log.Fields{
				"correlation_id":      messageObj.UUID,
				"worker_id":           id,
				"error":               err.Error(),
				"request_cluster":     deploymentRequest.Cluster,
				"request_namespace":   deploymentRequest.Namespace,
				"request_application": deploymentRequest.Application,
				"request_version":     deploymentRequest.Version,
				"request_strategy":    deploymentRequest.Strategy,
			}).Error(`Worker unable to ping cluster`)

			// Job Failed
			now := time.Now()
			job.Status = model.JobFailed
			job.RunAt = &now
			job.Result = fmt.Sprintf("Worker unable to ping cluster, UUID %s", messageObj.UUID)
			db.UpdateJobByID(&job)
			continue
		}

		ok, err = cluster.Deploy(deploymentRequest)

		if !ok || err != nil {
			log.WithFields(log.Fields{
				"correlation_id":      messageObj.UUID,
				"worker_id":           id,
				"error":               err.Error(),
				"request_cluster":     deploymentRequest.Cluster,
				"request_namespace":   deploymentRequest.Namespace,
				"request_application": deploymentRequest.Application,
				"request_version":     deploymentRequest.Version,
				"request_strategy":    deploymentRequest.Strategy,
			}).Error(`Worker unable deploy`)

			// Job Failed
			now := time.Now()
			job.Status = model.JobFailed
			job.RunAt = &now
			job.Result = fmt.Sprintf("Failure during deployment, UUID %s", messageObj.UUID)
			db.UpdateJobByID(&job)
			continue
		}

		log.WithFields(log.Fields{
			"correlation_id":      messageObj.UUID,
			"worker_id":           id,
			"request_cluster":     deploymentRequest.Cluster,
			"request_namespace":   deploymentRequest.Namespace,
			"request_application": deploymentRequest.Application,
			"request_version":     deploymentRequest.Version,
			"request_strategy":    deploymentRequest.Strategy,
		}).Info(`Deployment finished successfully`)

		// Job Succeeded
		now := time.Now()
		job.Status = model.JobSuccess
		job.RunAt = &now
		job.Result = "Deployment finished successfully"
		db.UpdateJobByID(&job)
	}
}
