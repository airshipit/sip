<h1>SIPCluster API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#airship.airshipit.org%2fv1">airship.airshipit.org/v1</a>
</li>
</ul>
<h2 id="airship.airshipit.org/v1">airship.airshipit.org/v1</h2>
<p>Package v1 contains API Schema definitions for the airship v1 API group</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="airship.airshipit.org/v1.BMCOpts">BMCOpts
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.JumpHostService">JumpHostService</a>)
</p>
<p>BMCOpts contains options for BMC communication.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>proxy</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.JumpHostService">JumpHostService
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPClusterServices">SIPClusterServices</a>)
</p>
<p>JumpHostService is an infrastructure service type that represents the sub-cluster jump-host service.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>SIPClusterService</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterService">
SIPClusterService
</a>
</em>
</td>
<td>
<p>
(Members of <code>SIPClusterService</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>bmc</code><br>
<em>
<a href="#airship.airshipit.org/v1.BMCOpts">
BMCOpts
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>sshkey</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.NodeSet">NodeSet
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPClusterSpec">SIPClusterSpec</a>)
</p>
<p>NodeSet are the the list of Nodes objects workers,
or ControlPlane that define expectations
for  the Tenant Clusters
Includes artifacts to associate with each defined namespace
Such as :
- Roles for the Nodes
- Flavor for theh Nodes image
- Scheduling expectations
- Scale of the group of Nodes</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>vmFlavor</code><br>
<em>
string
</em>
</td>
<td>
<p>VMFlavor is essentially a Flavor label identifying the
type of Node that meets the construction reqirements</p>
</td>
</tr>
<tr>
<td>
<code>spreadTopology</code><br>
<em>
<a href="#airship.airshipit.org/v1.SpreadTopology">
SpreadTopology
</a>
</em>
</td>
<td>
<p>PlaceHolder until we define the real expected
Implementation
Scheduling define constraints that allow the SIP Scheduler
to identify the required BMH&rsquo;s to allow CAPI to build a cluster</p>
</td>
</tr>
<tr>
<td>
<code>count</code><br>
<em>
<a href="#airship.airshipit.org/v1.VMCount">
VMCount
</a>
</em>
</td>
<td>
<p>Count defines the scale expectations for the Nodes</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SIPCluster">SIPCluster
</h3>
<p>SIPCluster is the Schema for the sipclusters API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterSpec">
SIPClusterSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>clusterName</code><br>
<em>
string
</em>
</td>
<td>
<p>ClusterName is the name of the cluster to associate machines with</p>
</td>
</tr>
<tr>
<td>
<code>nodes</code><br>
<em>
<a href="#airship.airshipit.org/v1.NodeSet">
map[./pkg/api/v1.VMRole]./pkg/api/v1.NodeSet
</a>
</em>
</td>
<td>
<p>Nodes defines the set of nodes to schedule for each vm role.</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterServices">
SIPClusterServices
</a>
</em>
</td>
<td>
<p>Services defines the services that are deployed when a SIPCluster is provisioned.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterStatus">
SIPClusterStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SIPClusterService">SIPClusterService
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.JumpHostService">JumpHostService</a>, 
<a href="#airship.airshipit.org/v1.SIPClusterServices">SIPClusterServices</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>nodeLabels</code><br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>nodePort</code><br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>nodeInterfaceId</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>clusterIP</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SIPClusterServices">SIPClusterServices
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPClusterSpec">SIPClusterSpec</a>)
</p>
<p>SIPClusterServices defines the services that are deployed when a SIPCluster is provisioned.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>loadBalancer</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterService">
[]SIPClusterService
</a>
</em>
</td>
<td>
<p>LoadBalancer defines the sub-cluster load balancer services.</p>
</td>
</tr>
<tr>
<td>
<code>auth</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterService">
[]SIPClusterService
</a>
</em>
</td>
<td>
<p>Auth defines the sub-cluster authentication services.</p>
</td>
</tr>
<tr>
<td>
<code>jumpHost</code><br>
<em>
<a href="#airship.airshipit.org/v1.JumpHostService">
[]JumpHostService
</a>
</em>
</td>
<td>
<p>JumpHost defines the sub-cluster jump host services.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SIPClusterSpec">SIPClusterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPCluster">SIPCluster</a>)
</p>
<p>SIPClusterSpec defines the desired state of a SIPCluster</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>clusterName</code><br>
<em>
string
</em>
</td>
<td>
<p>ClusterName is the name of the cluster to associate machines with</p>
</td>
</tr>
<tr>
<td>
<code>nodes</code><br>
<em>
<a href="#airship.airshipit.org/v1.NodeSet">
map[./pkg/api/v1.VMRole]./pkg/api/v1.NodeSet
</a>
</em>
</td>
<td>
<p>Nodes defines the set of nodes to schedule for each vm role.</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="#airship.airshipit.org/v1.SIPClusterServices">
SIPClusterServices
</a>
</em>
</td>
<td>
<p>Services defines the services that are deployed when a SIPCluster is provisioned.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SIPClusterStatus">SIPClusterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPCluster">SIPCluster</a>)
</p>
<p>SIPClusterStatus defines the observed state of SIPCluster</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.SpreadTopology">SpreadTopology
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.NodeSet">NodeSet</a>)
</p>
<h3 id="airship.airshipit.org/v1.VMCount">VMCount
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.NodeSet">NodeSet</a>)
</p>
<p>VMCount</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>active</code><br>
<em>
int
</em>
</td>
<td>
<p>INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
Important: Run &ldquo;make&rdquo; to regenerate code after modifying this file</p>
</td>
</tr>
<tr>
<td>
<code>standby</code><br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.VMRole">VMRole
(<code>string</code> alias)</h3>
<p>VMRole defines the states the provisioner will report
the tenant has having.</p>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
