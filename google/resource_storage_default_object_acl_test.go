package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccStorageDefaultObjectAcl_basic(t *testing.T) {
	t.Parallel()

	bucketName := testBucketName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStorageDefaultObjectAclDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntityBasic1, roleEntityBasic2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic1),
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic2),
				),
			},
		},
	})
}

func TestAccStorageDefaultObjectAcl_upgrade(t *testing.T) {
	t.Parallel()

	bucketName := testBucketName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStorageDefaultObjectAclDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntityBasic1, roleEntityBasic2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic1),
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic2),
				),
			},

			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntityBasic2, roleEntityBasic3_owner),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic2),
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic3_owner),
				),
			},

			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasicDelete(bucketName, roleEntityBasic1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic1),
					testAccCheckGoogleStorageDefaultObjectAclDelete(bucketName, roleEntityBasic2),
					testAccCheckGoogleStorageDefaultObjectAclDelete(bucketName, roleEntityBasic3_reader),
				),
			},
		},
	})
}

func TestAccStorageDefaultObjectAcl_downgrade(t *testing.T) {
	t.Parallel()

	bucketName := testBucketName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStorageDefaultObjectAclDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntityBasic2, roleEntityBasic3_owner),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic2),
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic3_owner),
				),
			},

			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntityBasic2, roleEntityBasic3_reader),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic2),
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic3_reader),
				),
			},

			resource.TestStep{
				Config: testGoogleStorageDefaultObjectsAclBasicDelete(bucketName, roleEntityBasic1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleStorageDefaultObjectAcl(bucketName, roleEntityBasic1),
					testAccCheckGoogleStorageDefaultObjectAclDelete(bucketName, roleEntityBasic2),
					testAccCheckGoogleStorageDefaultObjectAclDelete(bucketName, roleEntityBasic3_reader),
				),
			},
		},
	})
}

// Test that we allow the API to reorder our role entities without perma-diffing.
func TestAccStorageDefaultObjectAcl_unordered(t *testing.T) {
	t.Parallel()

	bucketName := testBucketName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStorageDefaultObjectAclDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testGoogleStorageDefaultObjectAclUnordered(bucketName),
			},
		},
	})
}

func testAccCheckGoogleStorageDefaultObjectAcl(bucket, roleEntityS string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		roleEntity, _ := getRoleEntityPair(roleEntityS)
		config := testAccProvider.Meta().(*Config)

		res, err := config.clientStorage.DefaultObjectAccessControls.Get(bucket,
			roleEntity.Entity).Do()

		if err != nil {
			return fmt.Errorf("Error retrieving contents of storage default Acl for bucket %s: %s", bucket, err)
		}

		if res.Role != roleEntity.Role {
			return fmt.Errorf("Error, Role mismatch %s != %s", res.Role, roleEntity.Role)
		}

		return nil
	}
}

func testAccStorageDefaultObjectAclDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {

		if rs.Type != "google_storage_default_object_acl" {
			continue
		}

		bucket := rs.Primary.Attributes["bucket"]

		_, err := config.clientStorage.DefaultObjectAccessControls.List(bucket).Do()
		if err == nil {
			return fmt.Errorf("Default Storage Object Acl for bucket %s still exists", bucket)
		}
	}
	return nil
}

func testAccCheckGoogleStorageDefaultObjectAclDelete(bucket, roleEntityS string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		roleEntity, _ := getRoleEntityPair(roleEntityS)
		config := testAccProvider.Meta().(*Config)

		_, err := config.clientStorage.DefaultObjectAccessControls.Get(bucket, roleEntity.Entity).Do()

		if err != nil {
			return nil
		}

		return fmt.Errorf("Error, Object Default Acl Entity still exists %s for bucket %s",
			roleEntity.Entity, bucket)
	}
}

func testGoogleStorageDefaultObjectsAclBasicDelete(bucketName, roleEntity string) string {
	return fmt.Sprintf(`
resource "google_storage_bucket" "bucket" {
	name = "%s"
}

resource "google_storage_default_object_acl" "acl" {
	bucket = "${google_storage_bucket.bucket.name}"
	role_entity = ["%s"]
}
`, bucketName, roleEntity)
}

func testGoogleStorageDefaultObjectsAclBasic(bucketName, roleEntity1, roleEntity2 string) string {
	return fmt.Sprintf(`
resource "google_storage_bucket" "bucket" {
	name = "%s"
}

resource "google_storage_default_object_acl" "acl" {
	bucket = "${google_storage_bucket.bucket.name}"
	role_entity = ["%s", "%s"]
}
`, bucketName, roleEntity1, roleEntity2)
}

func testGoogleStorageDefaultObjectAclUnordered(bucketName string) string {
	return fmt.Sprintf(`
resource "google_storage_bucket" "bucket" {
  name = "%s"
}

resource "google_storage_default_object_acl" "acl" {
  bucket = "${google_storage_bucket.bucket.name}"
  role_entity = ["%s", "%s", "%s", "%s", "%s"]
}
`, bucketName, roleEntityBasic1, roleEntityViewers, roleEntityOwners, roleEntityBasic2, roleEntityEditors)
}
