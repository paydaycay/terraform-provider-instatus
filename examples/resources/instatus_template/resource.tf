# Manage example template.
resource "instatus_template" "example" {
  subdomain = "some-subdomain"
  page_id = "PAGE_ID"
  name = "example name"
  type = "INCIDENT"
  status = "INVESTIGATING"
  notify = true
  message = "example message"
  components = [
    {
      id = "COMPONENT_ID"
      status = "MAJOROUTAGE"
    }
  ]
}