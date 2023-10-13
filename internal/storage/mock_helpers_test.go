package storage_test

import (
	"github.com/containers/image/v5/docker/reference"
	istorage "github.com/containers/image/v5/storage"
	cstorage "github.com/containers/storage"
	"github.com/cri-o/cri-o/internal/mockutils"
	containerstoragemock "github.com/cri-o/cri-o/test/mocks/containerstorage"
	criostoragemock "github.com/cri-o/cri-o/test/mocks/criostorage"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
)

// containers/image/storage.storageReference.StringWithinTransport
func mockStorageReferenceStringWithinTransport(storeMock *containerstoragemock.MockStore) mockutils.MockSequence {
	return mockutils.InOrder(
		storeMock.EXPECT().GraphOptions().Return([]string{}),
		storeMock.EXPECT().GraphDriverName().Return(""),
		storeMock.EXPECT().GraphRoot().Return(""),
		storeMock.EXPECT().RunRoot().Return(""),
	)
}

// containers/image/storage.Transport.GetStoreImage
// expectedImageName must be in the fully normalized format (reference.Named.String())!
// resolvedImageID may be "" to simulate a missing image
func mockGetStoreImage(storeMock *containerstoragemock.MockStore, expectedImageName, resolvedImageID string) mockutils.MockSequence { //nolint:unparam
	if resolvedImageID == "" {
		return mockutils.InOrder(
			storeMock.EXPECT().Image(expectedImageName).Return(nil, cstorage.ErrImageUnknown),
			mockResolveImage(storeMock, expectedImageName, ""),
		)
	}
	return mockutils.InOrder(
		storeMock.EXPECT().Image(expectedImageName).
			Return(&cstorage.Image{ID: resolvedImageID, Names: []string{expectedImageName}}, nil),
	)
}

// containers/image/storage.ResolveReference
// expectedImageName must be in the fully normalized format (reference.Named.String())!
// resolvedImageID may be "" to simulate a missing image
func mockResolveReference(storeMock *containerstoragemock.MockStore, storageTransportMock *criostoragemock.MockStorageTransport, expectedImageName, expectedImageID, resolvedImageID string) mockutils.MockSequence {
	var namedRef reference.Named
	if expectedImageName != "" {
		nr, err := reference.ParseNormalizedNamed(expectedImageName)
		Expect(err).To(BeNil())
		namedRef = nr
	}
	expectedRef, err := istorage.Transport.NewStoreReference(storeMock, namedRef, expectedImageID)
	Expect(err).To(BeNil())
	if resolvedImageID == "" {
		return mockutils.InOrder(
			storageTransportMock.EXPECT().ResolveReference(expectedRef).
				Return(nil, nil, istorage.ErrNoSuchImage),
		)
	}
	resolvedRef, err := istorage.Transport.NewStoreReference(storeMock, namedRef, resolvedImageID)
	Expect(err).To(BeNil())
	return mockutils.InOrder(
		storageTransportMock.EXPECT().ResolveReference(expectedRef).
			Return(resolvedRef,
				&cstorage.Image{ID: resolvedImageID, Names: []string{expectedImageName}},
				nil),
	)
}

// containers/image/storage.storageReference.resolveImage
// expectedImageNameOrID, if a name, must be in the fully normalized format (reference.Named.String())!
// resolvedImageID may be "" to simulate a missing image
func mockResolveImage(storeMock *containerstoragemock.MockStore, expectedImageNameOrID, resolvedImageID string) mockutils.MockSequence {
	if resolvedImageID == "" {
		return mockutils.InOrder(
			storeMock.EXPECT().Image(expectedImageNameOrID).Return(nil, cstorage.ErrImageUnknown),
			// Assuming expectedImageNameOrID does not have a digest, so resolveName does not call ImagesByDigest
			mockStorageReferenceStringWithinTransport(storeMock),
			mockStorageReferenceStringWithinTransport(storeMock),
		)
	}
	return mockutils.InOrder(
		storeMock.EXPECT().Image(expectedImageNameOrID).
			Return(&cstorage.Image{ID: resolvedImageID, Names: []string{expectedImageNameOrID}}, nil),
	)
}

// containers/image/storage.storageImageSource.getSize
func mockStorageImageSourceGetSize(storeMock *containerstoragemock.MockStore) mockutils.MockSequence {
	return mockutils.InOrder(
		storeMock.EXPECT().ListImageBigData(gomock.Any()).
			Return([]string{""}, nil), // A single entry
		storeMock.EXPECT().ImageBigDataSize(gomock.Any(), gomock.Any()).
			Return(int64(0), nil),
		// FIXME: This should also walk through the layer list and call store.Layer() on each, but we would have to mock the whole layer list.
	)
}

// containers/image/storage.storageReference.newImage
func mockNewImage(storeMock *containerstoragemock.MockStore, expectedImageName, resolvedImageID string) mockutils.MockSequence {
	return mockutils.InOrder(
		mockResolveImage(storeMock, expectedImageName, resolvedImageID),
		storeMock.EXPECT().ImageBigData(gomock.Any(), gomock.Any()).
			Return(testManifest, nil),
		mockStorageImageSourceGetSize(storeMock),
	)
}
