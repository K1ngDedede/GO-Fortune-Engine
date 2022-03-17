<script setup lang="ts">
import { onMounted, ref, watchEffect, type Ref } from 'vue'

const API_URL = "http://localhost:25566/"

const chandata = ref(null)
const clients = ref(null)
const files = ref(null)
const chan = ref('')
const autoUpdate = ref('')

watchEffect(async () => {
  fetchData()
  setInterval(fetchData,10000)  
})

async function fetchData(){
  chandata.value = await (await fetch(API_URL+"chan")).json()
  clients.value = await (await fetch(API_URL+"clients")).json()
  files.value = await (await fetch(API_URL+"files")).json()
  
}

function nochan(channel: string){
  if(chandata.value && Object.keys(chandata.value[channel]).length !=0){
    return false
  }else{
    return true
  }
}


function onlineClients(){
  let c = []
  let count = 0
  for (let y of Object.keys(clients.value!)){
    if(clients.value![y]["Online"]){
      c[count] = y
      count++
    }    
  }
  return c
}

function nofiles(){
  let c: null | any = files.value
  if(c && c.length!=0){
    return false
  }else{
    return true
  }
}

</script>

<template>
  <header>
    <h1>GO Fortune Engine</h1> 
    
  </header>
  <div>

  </div>
  <div>
    <p>Currently online: {{onlineClients().length}}</p>
    <br>
    <br>
    <br>
    <h2>File Log</h2>
      <table v-if="!nofiles()">
        <tr>
          <th>
            Filename
          </th>
          <th>
            Size
          </th>
          <th>
            Channel
          </th>
          <th>
            Sender
          </th>
          <th>
            Recipients
          </th>
          <th>
            Timestamp
          </th>
        </tr>
        <tr v-for="y in files">
          <td>
            {{y["Fname"]}}
          </td>
          <td>
            {{y["Fsize"]}}
          </td>
          <td>
            {{y["Channel"]}}
          </td>
          <td>
            {{y["Sender"]}}
          </td>
          <td>
            {{y["Recipients"]}}
          </td>
          <td>
            {{y["Tstamp"]}}
          </td>
        </tr>
      </table>
      <div v-else>
        No files have been transferred.
      </div>
  </div>
  
  <main class="debug">    
    <div>
      <h2>Online Clients</h2>
      <div v-if="onlineClients().length!=0">
        {{onlineClients()}}
      </div>
      <div v-else>
        No clients online.
      </div>
      <br>
      <br>
      <br>
      <h2>Open Channels</h2>
      <select v-model="chan" multiple style="width:100%">
        <option v-for="(value, name) in chandata">{{name}}</option>
      </select>
      <div v-if="chan">
        <div v-if="nochan(chan)">
          No clients in channel.
        </div>
        <div v-else>
          <h4>Clients in {{ chan }}</h4> 
          <ul>
            <li v-for="value in chandata![chan]">{{value["Cname"]}} @ {{value["Address"]}}</li>
          </ul>
        </div> 
      </div>
    </div>
  </main>
</template>

<style>
@import './assets/base.css';

#app {
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem;
  font-weight: normal;
  color: #DEDFD7;
}

table {
  border-collapse: collapse;
  width: 100%;
}

h1, h2, h3, h4, h5, h6{
  color: #EACD8E;
}


td, th {
  border: 1px solid #7E92A1;
  text-align: left;
  padding: 8px;
}

tr:nth-child(even) {
  background-color: #7E92A1;
}

header {
  line-height: 1.5;
}

.debug{
  
  margin-left: 20%;
}



@media (hover: hover) {
  a:hover {
    background-color: hsla(160, 100%, 37%, 0.2);
  }
}

@media (min-width: 1024px) {
  body {
    display: flex;
    place-items: center;
    background-color: #4D6879;
  }

  #app {
    display: grid;
    grid-template-columns: 1fr 1fr;
    padding: 0 2rem;
  }

  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }

  .logo {
    margin: 0 2rem 0 0;
  }
}
</style>
