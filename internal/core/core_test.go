package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) {
	if err := os.RemoveAll(".mygitserver"); err != nil {
		t.Fatalf("Failed to clean up test environment: %v", err)
	}

	InitializeRepository()
}

func cleanupTestRepo(t *testing.T) {
	if err := os.RemoveAll(".mygitserver"); err != nil {
		t.Fatalf("Failed to clean up after test: %v", err)
	}
}

func TestInitializeRepository(t *testing.T) {
	setupTestRepo(t)
	defer cleanupTestRepo(t)

	if _, err := os.Stat(".mygitserver"); os.IsNotExist(err) {
		t.Fatalf(".mygitserver directory not created")
	}

	headPath := filepath.Join(".mygitserver", "HEAD")
	if _, err := os.Stat(headPath); os.IsNotExist(err) {
		t.Fatalf("HEAD not created in .mygitserver")
	}

	refsPath := filepath.Join(".mygitserver", "refs", "heads")
	if _, err := os.Stat(refsPath); os.IsNotExist(err) {
		t.Fatalf("refs/heads not created in .mygitserver")
	}

	objectsPath := filepath.Join(".mygitserver", "objects")
	if _, err := os.Stat(objectsPath); os.IsNotExist(err) {
		t.Fatalf("objects directory not created in .mygitserver")
	}
}

func TestAddFile(t *testing.T) {
	setupTestRepo(t)
	defer cleanupTestRepo(t)

	testFileName := "testfile.txt"
	if err := ioutil.WriteFile(testFileName, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFileName)

	AddFile([]string{testFileName})

	stagedFiles, err := ioutil.ReadDir(filepath.Join(".mygitserver", "objects"))
	if err != nil {
		t.Fatalf("Failed to read objects directory: %v", err)
	}

	if len(stagedFiles) == 0 {
		t.Fatalf("No objects were created after adding file")
	}
}

func TestCommitChanges(t *testing.T) {
	setupTestRepo(t)
	defer cleanupTestRepo(t)

	testFileName := "testfile.txt"
	if err := ioutil.WriteFile(testFileName, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFileName)

	AddFile([]string{testFileName})

	CommitChanges([]string{"Initial commit"})

	commitFiles, err := ioutil.ReadDir(filepath.Join(".mygitserver", "objects"))
	if err != nil {
		t.Fatalf("Failed to read objects directory: %v", err)
	}

	if len(commitFiles) == 0 {
		t.Fatalf("No commit objects were created after commit")
	}

	headFilePath := filepath.Join(".mygitserver", "refs", "heads", "main")
	headContent, err := ioutil.ReadFile(headFilePath)
	if err != nil {
		t.Fatalf("Failed to read main branch file: %v", err)
	}

	if len(headContent) == 0 {
		t.Fatalf("No commit hash found in the main branch after commit")
	}
}

func TestCreateBranch(t *testing.T) {
	setupTestRepo(t)
	defer cleanupTestRepo(t)

	testFileName := "testfile.txt"
	if err := ioutil.WriteFile(testFileName, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFileName)

	AddFile([]string{testFileName})
	CommitChanges([]string{"Initial commit"})

	CreateBranch("feature")

	branchFilePath := filepath.Join(".mygitserver", "refs", "heads", "feature")
	if _, err := os.Stat(branchFilePath); os.IsNotExist(err) {
		t.Fatalf("Branch 'feature' was not created")
	}

	mainBranchHash, _ := ioutil.ReadFile(filepath.Join(".mygitserver", "refs", "heads", "main"))
	featureBranchHash, _ := ioutil.ReadFile(branchFilePath)
	if string(mainBranchHash) != string(featureBranchHash) {
		t.Fatalf("Feature branch commit hash does not match main branch commit hash")
	}
}

func TestMergeBranch(t *testing.T) {
	setupTestRepo(t)
	defer cleanupTestRepo(t)

	testFileName := "testfile.txt"
	if err := ioutil.WriteFile(testFileName, []byte("main branch content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFileName)

	AddFile([]string{testFileName})
	CommitChanges([]string{"Initial commit on main branch"})

	CreateBranch("feature")
	SwitchBranch("feature")

	if err := ioutil.WriteFile(testFileName, []byte("feature branch content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file in feature branch: %v", err)
	}

	AddFile([]string{testFileName})
	CommitChanges([]string{"Feature branch commit"})

	SwitchBranch("main")
	MergeBranch("feature")

	mainBranchHash, err := ioutil.ReadFile(filepath.Join(".mygitserver", "refs", "heads", "main"))
	if err != nil {
		t.Fatalf("Failed to read main branch after merge: %v", err)
	}

	if len(mainBranchHash) == 0 {
		t.Fatalf("Main branch was not updated with the merge commit")
	}
}
