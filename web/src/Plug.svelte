<script>

    export let plug;

</script>

<div class="plug" class:available="{plug.Is_available}">
    <ul>
        <li>
            Mac: {plug.Mac}
        </li>
        <li>
            Name {plug.Hostname}
        </li>
        <li>
            {plug.AddrV4}
        </li>
        <li>
            Last Seen: {plug.LastSeen}
        </li>
        <li>
            Available: {plug.Is_available}
        </li>
    </ul>
</div>

<style>

	div.plug {
		border: 1px red solid;
        margin: 10px;
	}

    div.available {
        background-color: cadetblue;
    }


</style>
