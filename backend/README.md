# AMBUSH Backend Breakdown

AMBUSH concerns society membership management. Warchest and Chaos need
information on society memberships, and a lot of these relations will be
queried straight from the central SpiceDB instance we all look to. However,
certain things are not (and should not be) owned by these client apps, and
should ask AMBUSH these things instead. This being executive membership
management, and the actual metadata of these memberships.

Thus, AMBUSH owns two domains in particular:

**1. Joint Executive Role Keeper (JERK)**

Frontend-facing API for society executive membership management, expected to be
used by the admin of AMBUSH. Handles the assignment of new executives to this
role, removal of the old team, etc.

Protocol: HTTP

**2. Federated Association & Profiling Service (FAPS)**

API for client apps to determine metadata about the various IDs they retrieve
from SpiceDB, such as organisation name, role name, membership year
etc.

Protocol: gRPC


