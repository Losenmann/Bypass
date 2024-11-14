:global funcBypassAddPG do={
:local url "https://postgrest.local/tb_bypass";
:local key "PostgREST_TOKEN";
:local scheme "YOU_SCHEME";
:local data;
:local input [:toarray $1]; 
:local cat $2;
:local res $3;

:local main do={
    :local data;
    :foreach k,i in=$1 do={
        if ($7 = "false") do={
            :set data ("{\"address\":\"".$i."\",\"category\":$2,\"resource\":$3}");
            :do {
                /tool/fetch url="$4" \
                    mode=https \
                    http-method=post \
                    http-header-field="Authorization: Bearer $5, Content-Type: application/json, Content-Profile: $6, Prefer: missing=default" \
                    http-data="$data" \
                    output=none;
                /log/info message="[BYPASS] Add address($i) Successful";
            } on-error={
                /log/error message="[BYPASS] Add address($i) Error";
            };
        } else {
            if ($k > 0) do={
                :set data ($data.",{\"address\":\"".$i."\",\"category\":$2,\"resource\":$3}");
            } else {
                :set data ("{\"address\":\"".$i."\",\"category\":$2,\"resource\":$3}");
            };
            if ( [:len $1] = $k+1) do={
                :return ("[".$data."]");
            };
        };
    };
};

:set data [$main $input $cat $res $url $key $scheme true];
:if ([:len $data] > 2) do={
    :do {
        /tool/fetch url="$url" \
            mode=https \
            http-method=post \
            http-header-field="Authorization: Bearer $key, Content-Type: application/json, Content-Profile: $scheme, Prefer: missing=default" \
            http-data="$data" \
            output=none;
        /log/info message="[BYPASS] Add address(all) Successful";
    } on-error={
        /log/error message="[BYPASS] Add address(all) Error";
        $main $input $cat $res $url $key $scheme false
    };
};
}

:global funcBypassSearch do={
:local name "googlevideo|youtube,instagram"
:local cat "9,2"
:local res "10,2"
:set name [:toarray $name]
:set cat [:toarray $cat]
:set res [:toarray $res]

:local search do={
    :global funcBypassAddPG;
    :local list "bypass";
    :local data;
    :local addr;
    :foreach i in=[/ip/dns/cache/find name~"$1"] do={
        :set addr [/ip/dns/cache/get $i name];
        :do {
            /ip/firewall/address-list/add list=$list address=$addr comment="$2; $3";
            :set data ($data.",".$addr);
        } on-error={};
    };
	:do {
        :set data [:pick $data 1 [:len $data]];
        $funcBypassAddPG $data $2 $3;
	} on-error={};
};

:foreach k,a in=$name do={
    $search $a [:pick $cat $k] [:pick $res $k]
};
}
