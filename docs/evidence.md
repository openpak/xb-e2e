# Evidence: xb-e2e

In-process end-to-end test of the native original-Xbox sign-in slice: provisioning
and enrolment in xb-account (fake core), MACS/AS/TGS in xb-auth, SG-init admission
in xb-gateway, over real sockets and mutual TLS. The console side is a test client
written against the same wire structures; no retail console has run it.
