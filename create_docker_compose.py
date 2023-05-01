import os
from configparser import ConfigParser

config = ConfigParser(os.environ)
config.read("config.ini")


data_dropper_instances = int(config["DEFAULT"]["DATA_DROPPER_INSTANCES"])
duration_averager_instances = int(config["DEFAULT"]["DURATION_AVERAGER_INSTANCES"])
precipitation_filter_instances = int(config["DEFAULT"]["PRECIPITATION_FILTER_INSTANCES"])
weather_joiner_instances = int(config["DEFAULT"]["WEATHER_JOINER_INSTANCES"])

stations_joiner_instances = 1

data_dropper_string = ""
for i in range(0, data_dropper_instances):
    data_dropper_string = data_dropper_string + f'''
  data_dropper_{i}:
    container_name: data_dropper_{i}
    environment:
      - ID={i}
      - PREV_STAGE_INSTANCES=1
      - WEATHER_JOINER_INSTANCES={weather_joiner_instances}
      - STATIONS_JOINER_INSTANCES={stations_joiner_instances}
    build:
      context: .
      dockerfile: ./data_dropper/Dockerfile
    entrypoint: /data_dropper
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy
'''


weather_joiner_string = ""
for i in range(0, weather_joiner_instances):
    weather_joiner_string = weather_joiner_string + f'''
  weather_joiner_{i}:
    container_name: weather_joiner_{i}
    environment:
      - ID={i}
      - PREV_STAGE_INSTANCES={data_dropper_instances}
      - NEXT_STAGE_INSTANCES={precipitation_filter_instances}
    build:
      context: .
      dockerfile: ./weather_joiner/Dockerfile
    entrypoint: /weather_joiner
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy
''' 


precipitation_filter_string = ""
for i in range(0, precipitation_filter_instances):
    precipitation_filter_string = precipitation_filter_string + f'''
  precipitation_filter_{i}:
    container_name: precipitation_filter_{i}
    environment:
      - ID={i}
      - PREV_STAGE_INSTANCES={weather_joiner_instances}
      - NEXT_STAGE_INSTANCES={duration_averager_instances}
    build:
      context: .
      dockerfile: ./precipitation_filter/Dockerfile
    entrypoint: /precipitation_filter
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy
'''   

duration_averager_string = ""
for i in range(0, duration_averager_instances):
    duration_averager_string = duration_averager_string + f'''
  duration_averager_{i}:
    container_name: duration_averager_{i}
    environment:
      - ID={i}
      - PREV_STAGE_INSTANCES={precipitation_filter_instances}
      - NEXT_STAGE_INSTANCES=1
    build:
      context: .
      dockerfile: ./duration_averager/Dockerfile
    entrypoint: /duration_averager
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy
'''   


file_content = f'''services:
  rabbitmq:
    container_name: rabbitmq
    build:
      context: ./rabbitmq
    ports:
      - 5672:5672
      - 15672:15672
    healthcheck:
      test: [ "CMD", "curl", "-f", "http://localhost:15672" ]
      interval: 10s
      timeout: 5s
      retries: 10

  client_handler:
    container_name: client_handler
    environment:
      - NEXT_STAGE_INSTANCES={data_dropper_instances}
    build:
      context: .
      dockerfile: ./client_handler/Dockerfile
    entrypoint: /client_handler
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy

  client:
    container_name: client
    build:
      context: .
      dockerfile: ./client/Dockerfile
    entrypoint: /client
    restart: on-failure
    depends_on:
      - client_handler
    volumes:
      - type: bind
        source: ./data
        target: /data
      - type: bind
        source: ./client/config.yaml
        target: /config.yaml

  duration_merger:
    container_name: duration_merger
    environment:
      - PREV_STAGE_INSTANCES={duration_averager_instances}
      - NEXT_STAGE_INSTANCES=1
    build:
      context: .
      dockerfile: ./duration_merger/Dockerfile
    entrypoint: /duration_merger
    restart: on-failure
    depends_on:
      rabbitmq:
        condition: service_healthy

{duration_averager_string}        
{precipitation_filter_string}   
{weather_joiner_string}   
{data_dropper_string}   
'''

f = open("compose.yaml", "w")
f.write(file_content)
f.close()