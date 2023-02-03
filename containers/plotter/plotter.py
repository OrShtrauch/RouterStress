import os
import glob
import json
import pandas as pd
from pandas import DataFrame
from datetime import datetime, timedelta

PATH = "/var/tmp/stress/data/"
DT_FORMAT = os.environ["dt_format"]
RUN_INDEX = 0

try:
    RUN_INDEX = int(os.environ["run_index"])
except (KeyError, ValueError):
    pass


def run():
    files, router_file = get_all_csv_files(RUN_INDEX)

    if files and router_file:
        router_df = trim_router_log(files[0], router_file)
        print(router_df)
        plot_router_graph(router_df)


def plot_router_graph(router_df: DataFrame):
    router_df = remove_date_from_df(router_df)

    router_plot = router_df.plot(
        x="timestamp",
        y="cpu",
        color="blue",
        label="CPU Usage(%)",
        title=f"CPU(%)\nmean={router_df['cpu'].mean()}\nstd={router_df['cpu'].std()}",
    )

    router_plot.set_xlabel("Time (s)")
    router_plot.set_ylabel("CPU Usage (%)")

    figure = router_plot.get_figure()
    figure.set_size_inches(15, 8)
    figure.savefig(f"{PATH}/cpu_usage.png")


def remove_date_from_df(df):
    for index, row in df.iterrows():
        timestamp = row["timestamp"]

        df.at[index, "timestamp"] = timestamp.split("-")[-1]

    return df


def get_all_csv_files(run_index: int):
    files: list[str] = [file for file in os.listdir(PATH) if file.endswith(".csv")]

    router_file: str = [file for file in files if "router" in file][0]
    files.remove(router_file)

    return [file for file in files if str(run_index) in file.split("_")[-1]], router_file


def get_time_adjusted_router_df(sample_file: str, router_file: str):
    router_df: DataFrame = pd.read_csv(f"{PATH}/{router_file}", sep=",")

    # getting a sample file to change the router df timestamp timezone
    df: DataFrame = pd.read_csv(f"{PATH}/{sample_file}", sep=",")

    router_timestamp: str = router_df.at[0, "timestamp"]
    sample_timestamp: str = df.at[0, "timestamp"]

    diff: datetime = from_timestamp(sample_timestamp) - from_timestamp(router_timestamp)

    hour_diff: int = round(diff.total_seconds() / 3600)

    for index, row in router_df.iterrows():
        timestamp: datetime = from_timestamp(row["timestamp"]) + timedelta(
            hours=hour_diff
        )
        router_df.at[index, "timestamp"] = timestamp.strftime(DT_FORMAT)

    return router_df

def trim_router_log(sample_file, router_file):
    router_df: DataFrame = get_time_adjusted_router_df(sample_file, router_file)
    df: DataFrame = pd.read_csv(f"{PATH}/{sample_file}", sep=",")

    start_time = from_timestamp(df.at[0, "timestamp"])
    end_time = from_timestamp(df.iloc[-1]["timestamp"])

    for index, row in router_df.iterrows():
        row_time = from_timestamp(row["timestamp"])

        if end_time < row_time or row_time < start_time:
            router_df.drop(index=index, inplace=True)

    return router_df

def from_timestamp(timestamp: str):
    return datetime.strptime(timestamp, DT_FORMAT)


if __name__ == "__main__":
    run()
