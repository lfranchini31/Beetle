// Copyright 2020 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"net/http"

	"github.com/clivern/beetle/internal/app/module"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetJob controller
func GetJob(c *gin.Context) {
	uuid := c.Param("uuid")

	db := module.Database{}

	err := db.AutoConnect()

	if err != nil {
		log.WithFields(log.Fields{
			"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
		}).Error(fmt.Sprintf(`Error: %s`, err.Error()))

		c.Status(http.StatusInternalServerError)
		return
	}

	defer db.Close()

	job := db.GetJobByUUID(uuid)

	if job.ID < 1 {
		log.WithFields(log.Fields{
			"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
		}).Info(fmt.Sprintf(`Job with UUID %s not found`, uuid))

		c.Status(http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
	}).Info(fmt.Sprintf(`Retrieve a job with UUID %s`, uuid))

	c.JSON(http.StatusOK, gin.H{
		"id":        job.ID,
		"uuid":      job.UUID,
		"status":    job.Status,
		"type":      job.Type,
		"runAt":     job.RunAt,
		"createdAt": job.CreatedAt,
		"updatedAt": job.UpdatedAt,
	})
}

// DeleteJob controller
func DeleteJob(c *gin.Context) {
	uuid := c.Param("uuid")

	db := module.Database{}

	err := db.AutoConnect()

	if err != nil {
		log.WithFields(log.Fields{
			"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
		}).Error(fmt.Sprintf(`Error: %s`, err.Error()))

		c.Status(http.StatusInternalServerError)
		return
	}

	defer db.Close()

	job := db.GetJobByUUID(uuid)

	if job.ID < 1 {
		log.WithFields(log.Fields{
			"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
		}).Info(fmt.Sprintf(`Job with UUID %s not found`, uuid))

		c.Status(http.StatusNotFound)
		return
	}

	log.WithFields(log.Fields{
		"CorrelationId": c.Request.Header.Get("X-Correlation-ID"),
	}).Info(fmt.Sprintf(`Deleting a job with UUID %s`, uuid))

	db.DeleteJobByID(job.ID)

	c.Status(http.StatusNoContent)
	return
}
