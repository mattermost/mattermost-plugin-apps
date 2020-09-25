# Proof Of Concept - Mattermost Apps

This plugin is being developed to test some concepts of creating Apps, which do not rely on a Go executable being installed on the Mattermost server/cluster to extend functionality.  The Apps will not be able to use Go RPC to communicate with the Mattermost Server, only through the "App Plugin" which acts as a sort of proxy to the server's activity.  

Apps will generally be communicating with our REST API and authenticating via OAuth. 

This is a precursor to our "Mattermost Apps" and "Mattermost Apps Marketplace" we are currently researching.  

## Getting Started

Join the "Integrations and Apps" channel to provide thoughts and feedback. 

## Contacts 

Dev: Lev Brouk (@lev.brouk)
PM: Aaron Rothschild (@aaron.rothschild)
