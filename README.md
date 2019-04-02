# Helios - Zero Trust proxy

## What is BeyondCorp?
Inside Google, BeyondCorp is core infrastructure that employees use every day.

Outside of Google, whether BeyondCorp is an architecture, a security philosophy, a positioning statement, a product you can just buy, a movement, or just common sense is hard to say.

To people who care about infrastructure security, BeyondCorp is the attacker never getting inside your VPN. To people who care about user experience (you know, users), BeyondCorp is never even having to use a VPN.

## BeyondCorp Vision
I hear from new people every day about how they’re trying to get there, and how they think about their current situations.

For many, BeyondCorp is just about learning, about developing their own perspectives as they help their organizations through a massive cloud shift. They may not be Google, but they can learn from Google.

For some especially large or ambitious organizations (you know who you are), BeyondCorp is something they’ve set out to build themselves. Some of these projects have been underway for a while. Some probably had a different name when they got started (“Redesign the VPN”), until the principal engineers convinced their leadership that a more modern approach was needed.

No matter who you are, the hardest questions seem to be:

* What do I need to build myself? (or, for some, “What do I get to build myself?”)
* Which, if any, parts of my current environments are coming with me? What will I be leaving behind? How do I keep the parts I want to keep?
* How do I make this change for all my applications? Or, for all my users?

These questions seem to be the barriers which stop people from being able to see how BeyondCorp could become practical for them. In these conversations, I’m mostly trying to help people develop their BeyondCorp vision, a way that makes sense for them to see the BeyondCorp concepts mapped into their specific reality. A lot of this is just listening about the individual problems they have, the things they’ve tried, and what they want to achieve. Then, finding out about the tools they have, and mapping those into the core reference architecture.

## The Architecture

One of the nice things about BeyondCorp, as opposed to “Zero Trust”, is that, instead of just telling you what you need to do without (no trust for you), BeyondCorp comes with a reference architecture: the platonic ideal of BeyondCorp. Whether individual organizations are implementing it in the same way as the original BeyondCorp architects at Google, or not, just doesn’t matter. They’re not wrong. They have the same goals. They have different (read: fewer) resources available, and some different problems.

There are a lot of ways to rethink your legacy security controls, and a lot of topics to rethink. The first three BeyondCorp papers from Google describe their RADIUS solution, and RADIUS is included in their first architecture diagram. Does that mean you need a RADIUS solution in order to be working on BeyondCorp at your company? Maybe not, right?

Here’s a BeyondCorp architecture diagram from Google:

![](https://www.scaleft.com/img/blogs/beyondcorp-outside/no-vpn-security-3-full.jpg)
> Pictured here are the components of BeyondCorp. Let’s look at each of them in turn and see how they fit in (maybe we’ll skip RADIUS).

### Access Proxy
The most significant component is the Access Proxy. From a whole-system design standpoint, the data needs of the Access Proxy will determine the properties of everything behind it: the Access Control Engine, the Pipeline, the Certificate Issuer. Google published an entire BeyondCorp paper on just this one component. This is also the only component which they’ve released as a product, the Identity Aware Proxy (IAP).

This most closely resembles a globally distributed, highly available, programmable load balancer. It enforces authorization decisions, and transports authorized user sessions to the internal services it protects (not pictured). This is not a place to pull your legacy WAM solution forward. It is a rich area for people who love to build software or infrastructure. The globally distributed nature of the Access Proxy is important for it to become a true VPN replacement and obsolete centralized VPN backhaul.

BeyondCorp centralizes policy enforcement in the Access Proxy. This replaces the old options of either centralizing policy enforcement by directly deferring all authentication mechanisms to a central Identity Provider (as in LDAP/PAM), or to distribute policy enforcement by serializing policies as collections of network rules, RSA keys, system configurations, account/password management solutions, etc.


### SSO, User/Group Database
I’m listing these two components together because they’re highly related. In most organizations, these are responsibilities of the corporate Identity Provider (IdP). Even though the same service will often be responsible for both functions, it’s valuable to consider them separately in the architecture. Consider how the Access Proxy is a refactoring of a single responsibility which legacy solutions often left as the responsibility of a central user database. BeyondCorp can be seen as a refactoring of the mechanics and responsibilities of AuthN and AuthZ away from user/group databases. Legacy user databases have never had the security, reliability, and scalability properties you want from critical infrastructure.

Organizations may use any of G Suite, Active Directory, Okta, SailPoint, LDAP, or more for these responsibilities. For International Friendship Day in 2016, at ScaleFT we even let users use Facebook logins for this (probably not a good long-term infrastructure commitment). As long as your organization is not using Facebook as a system of record, your IdP of choice can probably service both these responsibilities just fine. It’s something people can see easily bringing forward into their BeyondCorp deployments.

### Access Control Engine
The Access Control Engine is responsible for turning configurations and observations into authorization decisions. This is a green box in the diagram; a backend service that users and devices don’t directly interact with. This component is defined from its APIs: what it receives from the Pipeline and what it serves to the Access Proxy (in some places it is described as being inside the Access Proxy). There aren’t vendor products in this space. Not only that, there are no open standards yet which would enable you to build intercompatible software.

Thinking about these green box components is useful, not because it will necessarily reflect the structure of your own solution, but because it provides insight into the design decisions behind the BeyondCorp architecture. Since BeyondCorp is convergent evolution in action, there is a lot to learn about the problem space through reviewing the specific design choices Google made. These are important considerations for every organization beginning their BeyondCorp journey. Even with five BeyondCorp papers published, there will always be unknowns; so when people read through the papers and contemplate taking this journey themselves, there are many implementation-specific details and possibilities to consider.

For example, this design detail: Why have a separate Access Control Engine instead of just executing all policy logic in-line with handling a HTTP request in the Access Proxy?

* Access decisions can be memoized; within a specific time interval, a decision made over identical inputs should be invariant.
* The Access Proxy nodes only need active sessions in their working set. They do not need data on all resources, policies, users, or devices.
* Separating these services enables independent scaling in different regions; you will be more latency-sensitive in where you scale out your Access Proxy nodes since session data actually transits it; the Access Control Engine is less frequently called and lower throughput, so you would not need it in as many locations (distinction between data plane vs. control plane).

If you intend to design or assemble a similar system yourself, you’ll encounter these factors too.


### Pipeline
The Pipeline component (another green box) in the BeyondCorp diagram, more than anything else, shows that access control is inherently a distributed problem. The problem of access control is that your policies are in one place, the resources you want to protect are in another, and your users could be anywhere. To make an access decision, you need the truth about all the identities and properties of all three to be replicated/synchronized to a single place.

The Pipeline streams this data into the Access Proxy, but there’s not just one Access Proxy any more than there’s only one Pipeline. Both components are multi-region at Google scale. If you were to explode out this architecture diagram to show every individual cluster of each component, the whole diagram would be star-shaped, with the database components at the center, the Access Proxy nodes at the edge, and many Pipeline instances replicating data between them.

Having a highly distributed Pipeline and Access Proxy serves the end goal of getting security infrastructure out of the way of users; if the Proxy is going to be between the user and the resources they use, it should be as close to both as possible.

To replace the VPN with something better, you need an architecture which, like BeyondCorp, is capable of enforcing dynamic authorization decisions (based on information about the user and their devices), and responding intelligently to changes in configurations and real-world conditions.

Since, like a VPN, this solution will have a data plane which session data transits, latency and availability are critical. The Pipeline in Google’s BeyondCorp architecture shows, if not the exact architecture of every BeyondCorp solution, many of the distributed system properties that those solutions will need to solve for.

### Certificate Issuer
The Certificate Issuer issues certificates as authentication and authorization attestations for users and devices.

Certificates are the right choice for credential type to support BeyondCorp’s foundational concept of dynamic authorization.

For one reason, because a certificate is a time-limited cryptographically signed metadata object which may include many distinct attestations (metadata about identities, sessions, entitlements). Also, because the certificate is given to the user for use on their device, and presented by the user to the resource for verification at time of access (for example, when making a TLS connection).

This means the certificate can be seen as a message within a distributed system which is transmitted by the user. As long as the resource is available for the user to access, the authorization mechanism is also available.

### Trust Inference
Trust Inference is the process of assigning a trust tier to a device. At Google, trust tiers are organized in increasing levels of sensitivity, providing a spectrum of controls which can be applied to resources according to their security needs. As real world conditions (such as device configurations) change, those changes cascade to the trust tier, which can also change in real-time.

The Trust Inference component is another green box, this one adjacent to, but not connecting to, the Device Inventory DB. It’s interesting that it connects to the pipeline separately from the Device Inventory DB, and does not, for example, sit between them and annotate device inventory data with trust tiers.

The Trust Inference component also does not sit between the Device Inventory DB and the Certificate Issuer, as trust decisions are not made by the Certificate Issuer, and it only consumes device inventory data for use in certificate attributes.

### Device Inventory Database
No matter what your organization’s Device Inventory Database looks like, it has a role in your BeyondCorp solution. There are a huge number of vendor products out there already performing this function. Whether you’re using AirWatch, Fleetsmith, Jamf, Cisco ISE, or any of a thousand other solutions, the information in this database can be valuable for making authorization decisions based on policies. The exact ways in which devices are managed, and what compromises are made where, are highly organization-specific details.

Interestingly, though the Device Inventory DB outputs to the Pipeline, there is no line connecting the Managed Devices themselves to the Device Inventory DB, so that data from the real world device can enter the system. There are several related data flow questions to consider about the role of the Device Inventory Database in making authorization decisions.

* How frequently is data about devices updated in the inventory?
* How do changes in that data trigger policy re-evaluation?
* Do the Managed Device submit data directly to the Device Inventory DB, or is there a distributed data plane (such as the Pipeline) which connects those components?

### Managed Devices
There are no unmanaged devices in Google’s BeyondCorp diagram. The first BeyondCorp paper said that, at Google, “Only managed devices can access corporate applications.” But does that need to be true for all corporate applications, for all organizations?

Arguably, an unmanaged device has no place in a BeyondCorp world, since without information about a device, you can’t make a decision to trust that device. If you need to solve only for managed devices within your organization, that’s great. It’s hard enough, but there are many solutions available to choose from. What if you must solve for unmanaged devices though, what does that do to your BeyondCorp project?

From my point of view, the inevitability of unmanaged devices is not the end of the world. The real guarantees offered by most managed devices are not really that strong either. Until most managed devices have better hardware crypto features, the unknowns that threaten unmanaged devices are strong enough to threaten managed devices too.

It’s all academic at this point though. Unmanaged devices may not be in the diagram, but they are in our customers’ real worlds. BYOD isn’t a condition that our industry chose intentionally for its fantastic security properties, it’s one that happened to us.

At ScaleFT, we view this as a question of writing appropriate access policies. Some internal resources must have access restricted only to managed devices. Some internal resources may be assigned to a trust tier where information about the device is not a necessity. In that case, an unmanaged device may be fine.

The classic example of this is your company café’s lunch menu. You may not want to expose it for the entire internet to read everyday, but if someone views it on an unmanaged device, it’s not the end of the world. If that device is stolen and the attacker digs into the browser cache, later, they might find out what you had for lunch that day. It’s not your biggest risk. Lots of internal resources behind corporate VPNs are like this, and users understand intuitively that the security inconvenience is disproportionate to the risk.

The imposition of disproportionate inconvenience is what inspires users to find dangerous work-arounds for security solutions. Selecting and imposing appropriate policies for specific resources fixes this end-user adherence problem.

Good user experience enables better security, which is why the fifth BeyondCorp paper is titled, “The User Experience”. To achieve this user experience in a non-Google world, though, you may be solving for those users’ unmanaged devices too, even though they’re not in the diagram.

### The first rule of BeyondCorp
This might as well be the first rule of BeyondCorp: You can do it, but not exactly like Google. You’ll have to make the right tradeoffs for your organization, and discover the path forward for migrating your users, internal services, and use cases. Even with the release of Google’s IAP, BeyondCorp is just not something you can buy from Google. Interestingly, the challenges your organization faces with BYOD are structurally similar to the challenges Google would face trying to export their BeyondCorp implementation to other companies. There is a highly variable range of different problems which organizations face in implementing BeyondCorp. It is not one-size-fits-all.

This is a challenge we face at ScaleFT as well. For transformative change to be achievable, software and services need to meet companies and users where they are. If prescriptive mandates are too constraining, change will simply not happen. And, collectively, all organizations desperately need it to happen. Protecting internal services and data must be among our very highest priorities, and BeyondCorp is the best hope for this to have come along so far. That’s why I spend so much time with people just trying to develop that BeyondCorp vision with them. Every organization that begins this journey brings us one step closer to a BeyondCorp world.
