/*
Copyright 2018 Pusher Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"

	farosv1alpha1 "github.com/pusher/faros/pkg/apis/faros/v1alpha1"
	farosflags "github.com/pusher/faros/pkg/flags"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	farosGroupVersion = "faros.pusher.com/v1alpha1"
)

// OwnerIsNotClusterGitTrackPredicate filters events to check the owner of the event
// object is a ClusterGitTrack
type OwnerIsNotClusterGitTrackPredicate struct {
	client client.Client
}

// NewOwnerIsNotClusterGitTrackPredicate constructs a new OwnerIsNotClusterGitTrackPredicate
func NewOwnerIsNotClusterGitTrackPredicate(client client.Client) OwnerIsNotClusterGitTrackPredicate {
	return OwnerIsNotClusterGitTrackPredicate{
		client: client,
	}
}

// Create returns true if the event object owner is a ClusterGitTrack
func (p OwnerIsNotClusterGitTrackPredicate) Create(e event.CreateEvent) bool {
	return p.ownerIsNotClusterGitTrack(e.Meta.GetOwnerReferences())
}

// Update returns true if the event object owner is a ClusterGitTrack
func (p OwnerIsNotClusterGitTrackPredicate) Update(e event.UpdateEvent) bool {
	return p.ownerIsNotClusterGitTrack(e.MetaNew.GetOwnerReferences())
}

// Delete returns true if the event object owner is a ClusterGitTrack
func (p OwnerIsNotClusterGitTrackPredicate) Delete(e event.DeleteEvent) bool {
	return p.ownerIsNotClusterGitTrack(e.Meta.GetOwnerReferences())
}

// Generic returns true if the event object owner is a ClusterGitTrack
func (p OwnerIsNotClusterGitTrackPredicate) Generic(e event.GenericEvent) bool {
	return p.ownerIsNotClusterGitTrack(e.Meta.GetOwnerReferences())
}

func (p OwnerIsNotClusterGitTrackPredicate) ownerIsNotClusterGitTrack(ownerRefs []metav1.OwnerReference) bool {
	for _, ref := range ownerRefs {
		if ref.Kind == "ClusterGitTrack" && ref.APIVersion == farosGroupVersion {
			return false
		}
	}
	return true
}

// OwnerIsNotGitTrackPredicate filters events to check the owner of the event
// object is a GitTrack
type OwnerIsNotGitTrackPredicate struct {
	client client.Client
}

// NewOwnerIsNotGitTrackPredicate constructs a new OwnerIsNotGitTrackPredicate
func NewOwnerIsNotGitTrackPredicate(client client.Client) OwnerIsNotGitTrackPredicate {
	return OwnerIsNotGitTrackPredicate{
		client: client,
	}
}

// Create returns true if the event object owner is a GitTrack
func (p OwnerIsNotGitTrackPredicate) Create(e event.CreateEvent) bool {
	return p.ownerIsNotGitTrack(e.Meta.GetOwnerReferences())
}

// Update returns true if the event object owner is a GitTrack
func (p OwnerIsNotGitTrackPredicate) Update(e event.UpdateEvent) bool {
	return p.ownerIsNotGitTrack(e.MetaNew.GetOwnerReferences())
}

// Delete returns true if the event object owner is a GitTrack
func (p OwnerIsNotGitTrackPredicate) Delete(e event.DeleteEvent) bool {
	return p.ownerIsNotGitTrack(e.Meta.GetOwnerReferences())
}

// Generic returns true if the event object owner is a GitTrack
func (p OwnerIsNotGitTrackPredicate) Generic(e event.GenericEvent) bool {
	return p.ownerIsNotGitTrack(e.Meta.GetOwnerReferences())
}

func (p OwnerIsNotGitTrackPredicate) ownerIsNotGitTrack(ownerRefs []metav1.OwnerReference) bool {
	for _, ref := range ownerRefs {
		if ref.Kind == "GitTrack" && ref.APIVersion == farosGroupVersion {
			return false
		}
	}
	return true
}

// OurResponsibilityPredicate returns whether an event is our
// responsibility, based on the flags we're running with
type OurResponsibilityPredicate struct {
	client              client.Client
	gitTrackMode        farosflags.GitTrackMode
	clusterGitTrackMode farosflags.ClusterGitTrackMode
}

// NewOurResponsibilityPredicate constructs a new OurResponsibilityPredicate
func NewOurResponsibilityPredicate(client client.Client, gtmode farosflags.GitTrackMode, cgtmode farosflags.ClusterGitTrackMode) OurResponsibilityPredicate {
	return OurResponsibilityPredicate{
		client:              client,
		gitTrackMode:        gtmode,
		clusterGitTrackMode: cgtmode,
	}
}

func (p OurResponsibilityPredicate) isOurResponsibility(ownerRefs []metav1.OwnerReference) bool {
	gtoList := &farosv1alpha1.GitTrackObjectList{}
	err := p.client.List(context.TODO(), gtoList)
	if err != nil {
		// We can't list GTs so fail closed and ignore the requests
		return false
	}

	// build a set of uids
	gtoSet := make(map[types.UID]farosv1alpha1.GitTrackObject)
	for _, item := range gtoList.Items {
		gtoSet[item.UID] = item
	}

	for _, ref := range ownerRefs {
		// not a faros owner? not our problem
		if ref.APIVersion != farosGroupVersion {
			continue
		}
		// are we owned by a clustergittrackobject and are we handling ClusterGitTracks?
		if p.clusterGitTrackMode != farosflags.CGTMDisabled && ref.Kind == "ClusterGitTrackObject" {
			return true
		}

		if ref.Kind == "GitTrackObject" {
			// gtoSet contains all the gtos in our namespace, so check if we are owned by one of those.
			// TODO(dmo): we're assuming that we're in the same namespace here because any other construction
			// is invalid. Check if we need to be more proactive about invalid states
			gto, inSet := gtoSet[ref.UID]
			if p.gitTrackMode == farosflags.GTMEnabled && inSet {
				// make sure that we're owned by a gittrack and not a clustergittrack
				for _, gtref := range gto.GetOwnerReferences() {
					if gtref.APIVersion == farosGroupVersion && gtref.Kind == "GitTrack" {
						return true
					}
				}
			}

			// a ClusterGitTrack might have created this GitTrackObject
			if inSet && p.clusterGitTrackMode == farosflags.CGTMIncludeNamespaced {
				for _, gtref := range gto.GetOwnerReferences() {
					if gtref.APIVersion == farosGroupVersion && gtref.Kind == "ClusterGitTrack" {
						return true
					}
				}
			}
		}
	}
	return false
}

// Create returns true if the event object owner is our responsibility
func (p OurResponsibilityPredicate) Create(e event.CreateEvent) bool {
	return p.isOurResponsibility(e.Meta.GetOwnerReferences())
}

// Update returns true if the event object owner is our responsibility
func (p OurResponsibilityPredicate) Update(e event.UpdateEvent) bool {
	return p.isOurResponsibility(e.MetaNew.GetOwnerReferences())
}

// Delete returns true if the event object owner is our responsibility
func (p OurResponsibilityPredicate) Delete(e event.DeleteEvent) bool {
	return p.isOurResponsibility(e.Meta.GetOwnerReferences())
}

// Generic returns true if the event object owner is our responsibility
func (p OurResponsibilityPredicate) Generic(e event.GenericEvent) bool {
	return p.isOurResponsibility(e.Meta.GetOwnerReferences())
}
