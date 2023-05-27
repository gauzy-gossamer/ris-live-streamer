# Ris Live Streamer

BGP update streamer based on RIS Live (https://ris-live.ripe.net/).

Service reads BGP updates from Kafka in the format provided by pmacct (https://github.com/pmacct/pmacct) and streams to subscribers through websockets. Message format is based on RIS Live protocol, although not all features are implemented.
