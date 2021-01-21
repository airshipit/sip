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
<h3 id="airship.airshipit.org/v1.InfraConfig">InfraConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPClusterSpec">SIPClusterSpec</a>)
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
<code>serviceType</code><br>
<em>
<a href="#airship.airshipit.org/v1.InfraService">
InfraService
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>optional</code><br>
<em>
<a href="#airship.airshipit.org/v1.OptsConfig">
OptsConfig
</a>
</em>
</td>
<td>
</td>
</tr>
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
<code>nodelabels</code><br>
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
</tbody>
</table>
</div>
</div>
<h3 id="airship.airshipit.org/v1.InfraService">InfraService
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.InfraConfig">InfraConfig</a>)
</p>
<p>InfraService describes the type of infrastructure service that should be deployed when a sub-cluster is provisioned.</p>
<h3 id="airship.airshipit.org/v1.NodeSet">NodeSet
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.SIPClusterSpec">SIPClusterSpec</a>)
</p>
<p>NodeSet are the the list of Nodes objects workers,
or master that definee eexpectations
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
<code>vm-flavor</code><br>
<em>
string
</em>
</td>
<td>
<p>VMFlavor is  essentially a Flavor label identifying the
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
Scheduling define constraints the allows the SIP Scheduler
to identify the required  BMH&rsquo;s to allow CAPI to build a cluster</p>
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
<h3 id="airship.airshipit.org/v1.OptsConfig">OptsConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#airship.airshipit.org/v1.InfraConfig">InfraConfig</a>)
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
<code>sshkey</code><br>
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
<code>cluster-name</code><br>
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
map[./pkg/api/v1.VMRoles]./pkg/api/v1.NodeSet
</a>
</em>
</td>
<td>
<p>Nodes are the list of Nodes objects workers, or master that definee eexpectations
of the Tenant cluster
VMRole is either Control or Workers
VMRole VMRoles <code>json:&quot;vm-role,omitempty&quot;</code></p>
</td>
</tr>
<tr>
<td>
<code>infra</code><br>
<em>
<a href="#airship.airshipit.org/v1.InfraConfig">
[]InfraConfig
</a>
</em>
</td>
<td>
<p>InfraServices is a list of services that are deployed when a SIPCluster is provisioned.</p>
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
<code>cluster-name</code><br>
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
map[./pkg/api/v1.VMRoles]./pkg/api/v1.NodeSet
</a>
</em>
</td>
<td>
<p>Nodes are the list of Nodes objects workers, or master that definee eexpectations
of the Tenant cluster
VMRole is either Control or Workers
VMRole VMRoles <code>json:&quot;vm-role,omitempty&quot;</code></p>
</td>
</tr>
<tr>
<td>
<code>infra</code><br>
<em>
<a href="#airship.airshipit.org/v1.InfraConfig">
[]InfraConfig
</a>
</em>
</td>
<td>
<p>InfraServices is a list of services that are deployed when a SIPCluster is provisioned.</p>
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
<h3 id="airship.airshipit.org/v1.VMRoles">VMRoles
(<code>string</code> alias)</h3>
<p>VMRoles defines the states the provisioner will report
the tenant has having.</p>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
