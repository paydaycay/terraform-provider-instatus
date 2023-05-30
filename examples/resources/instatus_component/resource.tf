# Manage example component.
resource "instatus_component" "example" {
  page_id = "PAGE_ID"
  name = "App"
  status = "OPERATIONAL"
  show_uptime = true
  description = "Example App"
  order = 0
}