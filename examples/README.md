# Examples
This directory contains working Terraform configuration that is used for documentation.

The document generation tool looks for files in the following locations by default.
All other *.tf files besides the ones mentioned below are ignored by the documentation tool.

* **provider/provider.tf** example file for the provider index page
* **data-sources/`full data source name`/data-source.tf** example file for the named data source page
* **resources/`full resource name`/resource.tf** example file for the named resource page
